package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	Config "../config"
)

var config = Config.GetConfiguration()
var nodeExecutionDir = config.NodeExecutionDir
var addressesDir = config.AddressesDir
var sherlockFogDir = config.SherlockfogDir
var topologyCreationDir = config.TopologyCreationDir

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

var hostIps = [7][]byte{[]byte("10.1.10.119\n"), []byte("10.1.10.120\n"), []byte("10.1.10.137\n"), []byte("10.1.10.138\n"), []byte("10.1.10.162\n"), []byte("10.1.10.163\n"), []byte("10.1.10.166\n")}

func writeLineToFile(file *os.File, content string) {
	if _, err := file.Write([]byte(content+"\n")); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to %s.\n %s\n", file.Name(), err.Error()))
	}
}

func addSemaphore(scriptFile *os.File, nodes []btcNode) {

	writeLineToFile(scriptFile, "run n0 netns n0 sleep 30s")

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d /usr/local/go/bin/go run "+nodeExecutionDir+"/semaphore.go %d", nodes[i].Host, nodes[i].Host, nodes[i].Id))
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

func makeLogicalLayer(scriptFile *os.File, nodes []btcNode) map[int]bool {

	nodeIdToHost := make(map[int]int)
	hostHasNode := make(map[int]bool)

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash "+nodeExecutionDir+"/invokeBitcoin.sh %d -dificulta=0 -dbcache=3072",nodes[i].Host,nodes[i].Host,nodes[i].Id))//-loadblock=/home/mgeier/ndecarli/blk00000.dat -loadblock=/home/mgeier/ndecarli/blk00001.dat
		nodeIdToHost[nodes[i].Id] = nodes[i].Host
		hostHasNode[nodes[i].Host] = true
	}

	writeLineToFile(scriptFile, "")

	addSemaphore(scriptFile, nodes)

	for i:=0; i<len(nodes); i++ {
		for j:=0; j<len(nodes[i].ConnectedTo); j++{
			writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash "+nodeExecutionDir+"/connectNodes.sh %d %d n%d", nodes[i].Host, nodes[i].Host, nodes[i].Id, nodes[i].ConnectedTo[j], nodeIdToHost[nodes[i].ConnectedTo[j]]))
		}
	}

	return hostHasNode
}

func makeBlockChain(scriptFile *os.File, nodes []btcNode, hostHasNode map[int]bool) {

	writeLineToFile(scriptFile,fmt.Sprintf("\nrun n0 netns n0 bash "+nodeExecutionDir+"/invokeBitcoin.sh %d -dificulta=0 -dbcache=3072\n", len(nodes)))//-loadblock=/home/mgeier/ndecarli/blk00000.dat -loadblock=/home/mgeier/ndecarli/blk00001.dat

	addSemaphore(scriptFile, []btcNode{btcNode{Id: len(nodes), Host:0}})

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash "+nodeExecutionDir+"/connectNodes.sh %d %d n0", nodes[i].Host, nodes[i].Host, nodes[i].Id, len(nodes)))
		writeLineToFile(scriptFile,fmt.Sprintf("run n0 netns n0 bash "+nodeExecutionDir+"/connectNodes.sh %d %d n%d", len(nodes), nodes[i].Id, nodes[i].Host))
	}

	writeLineToFile(scriptFile, "")

	for i, j := 0, 0; i<len(nodes) && j<len(hostIps); i++ {
		if hostHasNode[i] {
			writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d rm -f "+addressesDir+"/addrN*", i, i))
			j++
		}
	}

	writeLineToFile(scriptFile, "")

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash "+nodeExecutionDir+"/bitcoindo.sh %d getnewaddress > "+addressesDir+"/addrN%d", nodes[i].Host, nodes[i].Host, nodes[i].Id, nodes[i].Id))
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash "+nodeExecutionDir+"/bitcoindo.sh %d getnewaddress >> "+addressesDir+"/addrN%d\n", nodes[i].Host, nodes[i].Host, nodes[i].Id, nodes[i].Id))

	}

	for i, j := 0, 0; i<len(nodes) && j<len(hostIps); i++ {
		if hostHasNode[i] {
			writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d scp -q -o StrictHostKeyChecking=no "+addressesDir+"/addrN* n0:"+addressesDir, i, i))//, nodes[i].Id, nodes[i].Id))
			j++
		}
	}

	writeLineToFile(scriptFile,fmt.Sprintf("\nrun n0 netns n0 /usr/local/go/bin/go run "+nodeExecutionDir+"/generateBlockchain.go %d\n", len(nodes)))

	addSemaphore(scriptFile, nodes)

	writeLineToFile(scriptFile,fmt.Sprintf("run n0 netns n0 bash "+nodeExecutionDir+"/bitcoindo.sh %d stop\n", len(nodes)))

	writeLineToFile(scriptFile, "\nrun n0 netns n0 sleep 1m")

	writeLineToFile(scriptFile, fmt.Sprintf("\nrun n0 netns n0 rm /root/btcCoreLogN%d", len(nodes)))

	writeLineToFile(scriptFile, fmt.Sprintf("\nrun n0 netns n0 rm -r /home/ndecarli/regtestData/%d", len(nodes)))
}

func startEngines(scriptFile *os.File, topology *GraphJson) {

	for i:=0; i<len(topology.BtcNodes); i++ {
		writeLineToFile(scriptFile, fmt.Sprintf("run n%d netns n%d /usr/local/go/bin/go run "+nodeExecutionDir+"/launcher.go /usr/local/go/bin/go run "+nodeExecutionDir+"/testEngine.go %d %f", topology.BtcNodes[i].Host, topology.BtcNodes[i].Host, topology.BtcNodes[i].Id, topology.BtcNodes[i].HashingPower))
		writeLineToFile(scriptFile, "\nrun n0 netns n0 sleep 1s")
	}

	writeLineToFile(scriptFile, "")

	for i:=1; i<topology.Network.Hosts; i++ {
		writeLineToFile(scriptFile, fmt.Sprintf("run n%d netns n%d /usr/local/go/bin/go run "+nodeExecutionDir+"/launcher.go /usr/local/go/bin/go run "+nodeExecutionDir+"/pingEngine.go %d %d", i, i, topology.Network.Hosts, i))
	}

	writeLineToFile(scriptFile, fmt.Sprintf("run n0 netns n0 /usr/local/go/bin/go run "+nodeExecutionDir+"/pingEngine.go %d 0", topology.Network.Hosts))
}

func teardown(scriptFile *os.File, nodes []btcNode) {

	writeLineToFile(scriptFile, "")

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash "+nodeExecutionDir+"/bitcoindo.sh %d stop", nodes[i].Host,nodes[i].Host,nodes[i].Id))
	}

	writeLineToFile(scriptFile, "\nrun n0 netns n0 sleep 1m")
}

func launchSherlockFog(scriptFile *os.File, numberOfHosts int, hostHasNode map[int]bool) {

	if fileErr := scriptFile.Close(); fileErr == nil {
		ipsFilename := os.Args[1]+"ips.txt"
		if ipsFile, err := os.Create(ipsFilename); err == nil {
			j := 0
			k := len(hostIps)-1
			for i:=0; i<numberOfHosts && err == nil; i++ {
				if hostHasNode[i] {
					_, err = ipsFile.Write(hostIps[k])
					if k>0 {
						k--
					} else {
						k = len(hostIps)-1
					}
				} else {
					_, err = ipsFile.Write(hostIps[j])
					j++
					j %= len(hostIps)
				}
			}
			if err != nil {
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

	hostHasNode := makeLogicalLayer(scriptFile, topology.BtcNodes)

	makeBlockChain(scriptFile, topology.BtcNodes, hostHasNode)

	startEngines(scriptFile, &topology)

	teardown(scriptFile, topology.BtcNodes)

	launchSherlockFog(scriptFile, topology.Network.Hosts, hostHasNode)
}
