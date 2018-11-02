package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
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
	HashingPower float64 `json:"HashingPower"`
	ConnectedTo []int `json:"connectedTo"`
}

func writeLineToFile(file *os.File, content string) {
	if _, err := file.Write([]byte(content+"\n")); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to file.\n %s\n", err.Error()))
	}
}

func createFile() (scriptFile *os.File) {
	var err error

	if scriptFile, err = os.Create(os.Args[1]+".fog"); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file.\n %s\n", err.Error()))
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

func makePhysicalLayer(scriptFile *os.File, topology *network) {

	writeLineToFile(scriptFile ,fmt.Sprintf("for i in 0..%d do", topology.Hosts))
	writeLineToFile(scriptFile ,"\tdef n{i}\nend for\n")

	for i := 0; i<len(topology.Connections); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("connect n%d n%d %dms",topology.Connections[i].A,topology.Connections[i].B,topology.Connections[i].Latency))
	}

	writeLineToFile(scriptFile, "\nbuild-network\n")
}

func makeLogicalLayer(scriptFile *os.File, nodes []btcNode, nodeIdToHost map[int]int) {

	for i:=0; i<len(nodes); i++ {
		writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash /home/mgeier/ndecarli/invokeBitcoin.sh %d -simuLambda=%f",nodes[i].Host,nodes[i].Host,nodes[i].Id,nodes[i].HashingPower))
		nodeIdToHost[nodes[i].Id] = nodes[i].Host
	}

	writeLineToFile(scriptFile, "")

	for i:=0; i<len(nodes); i++ {
		for j:=0; j<len(nodes[i].ConnectedTo); j++{
			writeLineToFile(scriptFile,fmt.Sprintf("run n%d netns n%d bash /home/mgeier/ndecarli/connectNodes.sh %d %d n%d", nodes[i].Host, nodes[i].Host, nodes[i].Id, nodes[i].ConnectedTo[j], nodeIdToHost[nodes[i].ConnectedTo[j]]))
		}
	}
}

func makeBlockChain(scriptFile *os.File, nodes []btcNode, nodeIdToHost map[int]int) {


}

func main(){

	if len(os.Args) < 3 {
		os.Stderr.WriteString("Missing Json filenames as arguments.\n")
		return
	}

	scriptFile := createFile()

	topology := parseJson()

	makePhysicalLayer(scriptFile, &topology.Network)

	nodeIdToHost := make(map[int]int)

	makeLogicalLayer(scriptFile, topology.BtcNodes, nodeIdToHost)

	makeBlockChain(scriptFile, topology.BtcNodes, nodeIdToHost)
}
