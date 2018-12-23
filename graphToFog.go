package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"os/exec"
	"bytes"
)

type GraphJson struct {
	Network network `json:"network"`
	BtcNodes []btcNode `json:"btcNodes"`
}

type network struct {
	Connections []NetworkConnection `json:"connections"`
	Hosts int `json:"hosts"`
}

type NetworkConnection struct {
	A int `json:"a"`
	B int `json:"b"`
	Latency int `json:"latencyMs"`
}

type btcNode struct {
	Id int `json:"id"`
	Host int  `json:"host"`
	HashingPower float64 `json:"hashingPower"`
	ConnectedTo []int `json:"connectedTo"`
}

var hostIps = [3][]byte{[]byte("10.1.10.162\n"), []byte("10.1.10.163\n"), []byte("10.1.10.166\n")}//, []byte("10.1.10.167\n")}

func writeLineToFile(file *os.File, content string) {
	if _, err := file.Write([]byte(content+"\n")); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to %s.\n %s\n", file.Name(), err.Error()))
	}
}

func addSemaphore(scriptFile *os.File, nodes []btcNode) {

	writeLineToFile(scriptFile, "run n0 netns n0 sleep 1m")

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d /usr/local/go/bin/go run /home/mgeier/ndecarli/semaphore.go %d", nodes[i].Host, nodes[i].Host, nodes[i].Id))
	}

	writeLineToFile(scriptFile, "")
}

func createFile() (scriptFile *os.File) {
	var err error

	if scriptFile, err = os.Create(os.Args[1]+".fog"); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file "+os.Args[1]+".fog\n %s\n", err.Error()))
	}

	return scriptFile
}

func parseJson() (topology GraphJson) {

	if jsonBytes, err := ioutil.ReadFile(os.Args[2]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse physical layer file.\n %s\n", err.Error()))
	} else if err = json.Unmarshal(jsonBytes, &topology); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse physical layer json.\n %s\n", err.Error()))
	}
	return
}

func delayString(latency int) string {
	if latency > 0 {
		return fmt.Sprintf(" %dms", latency)
	} else {
		return ""
	}
}

func makePhysicalLayer(scriptFile *os.File, networkTopology *network) {

	writeLineToFile(scriptFile ,fmt.Sprintf("for i in 0..%d do\n\tdef n{i}\nend for\n", networkTopology.Hosts))

	for i := 0; i<len(networkTopology.Connections); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("connect n%d n%d%s",networkTopology.Connections[i].A,networkTopology.Connections[i].B,delayString(networkTopology.Connections[i].Latency)))
	}

	writeLineToFile(scriptFile, "\nbuild-network\n")
}

func makeLogicalLayer(scriptFile *os.File, nodes []btcNode) {

	nodeIdToHost := make(map[int]int)

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash /home/mgeier/ndecarli/invokeBitcoin.sh %d -simuLambda=%f -dbcache=2048 -loadblock=/home/mgeier/ndecarli/blk00000.dat -loadblock=/home/mgeier/ndecarli/blk00001.dat",nodes[i].Host,nodes[i].Host,nodes[i].Id,nodes[i].HashingPower))
		nodeIdToHost[nodes[i].Id] = nodes[i].Host
	}

	writeLineToFile(scriptFile, "")

	addSemaphore(scriptFile, nodes)

	for i:=0; i<len(nodes); i++ {
		for j:=0; j<len(nodes[i].ConnectedTo); j++{
			writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash /home/mgeier/ndecarli/connectNodes.sh %d %d n%d", nodes[i].Host, nodes[i].Host, nodes[i].Id, nodes[i].ConnectedTo[j], nodeIdToHost[nodes[i].ConnectedTo[j]]))
		}
	}
}

func makeBlockChain(scriptFile *os.File, nodes []btcNode) {

	writeLineToFile(scriptFile,fmt.Sprintf("\nrun n0 netns n0 bash /home/mgeier/ndecarli/invokeBitcoin.sh %d -dificulta=0 -dbcache=2048 -loadblock=/home/mgeier/ndecarli/blk00000.dat -loadblock=/home/mgeier/ndecarli/blk00001.dat\n", len(nodes)))

	addSemaphore(scriptFile, []btcNode{btcNode{Id: len(nodes), Host:0}})

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash /home/mgeier/ndecarli/connectNodes.sh %d %d n0", nodes[i].Host, nodes[i].Host, nodes[i].Id, len(nodes)))
		writeLineToFile(scriptFile,fmt.Sprintf("run n0 netns n0 bash /home/mgeier/ndecarli/connectNodes.sh %d %d n%d", len(nodes), nodes[i].Id, nodes[i].Host))
	}

	writeLineToFile(scriptFile, "")

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash /home/mgeier/ndecarli/bitcoindo.sh %d getnewaddress > /home/mgeier/ndecarli/addrN%d", nodes[i].Host, nodes[i].Host, nodes[i].Id, nodes[i].Id))
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash /home/mgeier/ndecarli/bitcoindo.sh %d getnewaddress >> /home/mgeier/ndecarli/addrN%d", nodes[i].Host, nodes[i].Host, nodes[i].Id, nodes[i].Id))

		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d scp -q -o StrictHostKeyChecking=no /home/mgeier/ndecarli/addrN%d n0:/home/mgeier/ndecarli/addrN%d\n", nodes[i].Host, nodes[i].Host, nodes[i].Id, nodes[i].Id))
	}

	writeLineToFile(scriptFile,fmt.Sprintf("run n0 netns n0 /usr/local/go/bin/go run /home/mgeier/ndecarli/generateBlockchain.go %d\n", len(nodes)))

	addSemaphore(scriptFile, nodes)

	writeLineToFile(scriptFile,fmt.Sprintf("run n0 netns n0 bash /home/mgeier/ndecarli/bitcoindo.sh %d stop\n", len(nodes)))
}

func startEngines(scriptFile *os.File, topology *GraphJson) {

	for i:=0; i<len(topology.BtcNodes); i++ {
		writeLineToFile(scriptFile, fmt.Sprintf("run n%d netns n%d /usr/local/go/bin/go run /home/mgeier/ndecarli/launcher.go /usr/local/go/bin/go run /home/mgeier/ndecarli/testEngine.go %d", topology.BtcNodes[i].Host, topology.BtcNodes[i].Host, topology.BtcNodes[i].Id))
	}

	writeLineToFile(scriptFile, "")

	for i:=1; i<topology.Network.Hosts; i++ {
		writeLineToFile(scriptFile, fmt.Sprintf("run n%d netns n%d /usr/local/go/bin/go run /home/mgeier/ndecarli/launcher.go /usr/local/go/bin/go run /home/mgeier/ndecarli/pingEngine.go %d %d", i, i, topology.Network.Hosts, i))
	}

	writeLineToFile(scriptFile, fmt.Sprintf("run n0 netns n0 /usr/local/go/bin/go run /home/mgeier/ndecarli/pingEngine.go %d 0", topology.Network.Hosts))
}

func teardown(scriptFile *os.File, nodes []btcNode) {

	writeLineToFile(scriptFile, "")

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash /home/mgeier/ndecarli/bitcoindo.sh %d stop", nodes[i].Host,nodes[i].Host,nodes[i].Id))
	}

	writeLineToFile(scriptFile, "\nrun n0 netns n0 sleep 1m")
}

func launchSherlockFog(scriptFile *os.File, numberOfHosts int) {

	if fileErr := scriptFile.Close(); fileErr == nil {
		ipsFilename := os.Args[1]+"ips.txt"
		if ipsFile, err := os.Create(ipsFilename); err == nil {
			for i:=0; i<numberOfHosts && err == nil; i++ {
				_, err = ipsFile.Write(hostIps[i%len(hostIps)])
			}
			if err == nil {
				launchFog := exec.Command(/*"bash", "-c", */"python3", "/home/mgeier/repos/sherlockfog/sherlockfog.py", "/home/mgeier/ndecarli/"+scriptFile.Name(), "--real-host-list=/home/mgeier/ndecarli/"+ipsFilename, "--cpu-exclusive=False")//, "> sherlockOut")
				var stdErr bytes.Buffer
				launchFog.Stderr = &stdErr
				b, err := launchFog.Output()
				if err != nil {
					os.Stderr.WriteString(fmt.Sprintf("Failed to launch sherlock fog.\n%s : %s\n", err.Error(), stdErr.String()))
				}
				fmt.Println(string(b))
				fmt.Println(stdErr.String())
				/*launchFog := exec.Command("bash", "-c", "python3 /home/mgeier/repos/sherlockfog/sherlockfog.py /home/mgeier/ndecarli/"+scriptFile.Name()+" --real-host-list=/home/mgeier/ndecarli/"+ipsFilename+" --cpu-exclusive=False", "> sherlockOut")
				if err = launchFog.Start(); err != nil {
					os.Stderr.WriteString(fmt.Sprintf("Failed to launch sherlock fog.\n%s\n", err.Error()))
				}*/
			} else {
				os.Stderr.WriteString(fmt.Sprintf("Failed to write ips file.\n %s\n", err.Error()))
			}
		} else {
			os.Stderr.WriteString(fmt.Sprintf("Failed to create ips file.\n %s\n", err.Error()))
		}
	} else {
		os.Stderr.WriteString("Could not close fog file properly.\n")
	}
}

func main() {

	if len(os.Args) < 3 {
		os.Stderr.WriteString("Missing Json filenames as arguments.\n")
		return
	}

	scriptFile := createFile()

	topology := parseJson()

	makePhysicalLayer(scriptFile, &topology.Network)

	makeLogicalLayer(scriptFile, topology.BtcNodes)

	makeBlockChain(scriptFile, topology.BtcNodes)

	startEngines(scriptFile, &topology)

	teardown(scriptFile, topology.BtcNodes)

	launchSherlockFog(scriptFile, topology.Network.Hosts)

}
