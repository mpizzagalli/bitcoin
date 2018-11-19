package main

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"strconv"
)

type DistributionJson struct {
	CountryDistribution []BtcCountry `json:"country_distribution"`
	CountryLatency []CountryConnection `json:"country_latency"`
	Pools []BtcPool `json:"pool_distribution"`
}

type BtcCountry struct {
	Id string `json:"id"`
	NetworkShare float64 `json:"share"`
	InnerLatency int `json:"inner_latency"`
}

type CountryConnection struct {
	A string `json:"a"`
	B string `json:"b"`
	Latency int `json:"latency_ms"`
}

type BtcPool struct {
	Id string `json:"id"`
	Nodes []PoolNode `json:"nodes"`
	HPShare float64 `json:"network_share"`
}

type PoolNode struct {
	Country string `json:"country_id"`
	PoolShare float64 `json:"pool_share"`
}

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

func createFile() (scriptFile *os.File) {
	var err error

	if scriptFile, err = os.Create(os.Args[1]+".json"); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file "+os.Args[1]+".fog\n %s\n", err.Error()))
	}

	return scriptFile
}

func parseJson() (topology DistributionJson) {

	if jsonBytes, err := ioutil.ReadFile(os.Args[2]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse physical layer file.\n %s\n", err.Error()))
	} else if err = json.Unmarshal(jsonBytes, &topology); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse physical layer json.\n %s\n", err.Error()))
	}
	return
}

func parseAmountOfBtcNodes() int {

	var err error
	var btcNodes int

	if btcNodes, err = strconv.Atoi(os.Args[3]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file "+os.Args[1]+".fog\n %s\n", err.Error()))
	}

	return btcNodes
}

func generateNetwork(data *DistributionJson) (hostNetwork network, countryIdtoHostId map[string][2]int) {

	//Cantidad de hosts es cantidad de nodos + cantidad de paises
	hostNetwork.Hosts = parseAmountOfBtcNodes() + len(data.CountryDistribution)
	hostNetwork.Connections = make([]NetworkConnection, 0)

	var amountOfHosts float64 = float64(hostNetwork.Hosts)

	countryIdtoHostId = make(map[string][2]int)

	for i, j := 0, 0; i< len(data.CountryDistribution); i++ {
		nodesInCountry := int(amountOfHosts * data.CountryDistribution[i].NetworkShare + 0.5)
		countryIdtoHostId[data.CountryDistribution[i].Id] = [2]int{j, nodesInCountry}
		for k, c := j+1, j+nodesInCountry; k<=c; k++ {
			hostNetwork.Connections = append(hostNetwork.Connections, NetworkConnection{j, k, data.CountryDistribution[i].InnerLatency})
		}

		j += nodesInCountry + 1
	}

	for i:=0; i<len(data.CountryLatency); i++ {

		hostA:=countryIdtoHostId[data.CountryLatency[i].A][0]
		hostB:=countryIdtoHostId[data.CountryLatency[i].B][0]

		hostNetwork.Connections = append(hostNetwork.Connections, NetworkConnection{hostA, hostB, data.CountryLatency[i].Latency})
	}

	return
}

func generateNodes(data *DistributionJson, countryIdtoHostId map[string][2]int) []btcNode {
	return nil
}

func writeTopology(hostNetwork network, nodes []btcNode) {

	jsonData := GraphJson{Network:hostNetwork, BtcNodes:nodes}

	jsonBytes, _ := json.Marshal(jsonData)

	jsonFile := createFile()

	if _, err := jsonFile.Write(jsonBytes); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to %s.\n %s\n", jsonFile.Name(), err.Error()))
	}
}

func main(){

	if len(os.Args) < 4 {
		os.Stderr.WriteString("Missing Json filenames and total amount of nodes as arguments.\n")
		return
	}

	data := parseJson()

	network, countryIdtoHostId := generateNetwork(&data)

	nodes := generateNodes(&data, countryIdtoHostId)

	writeTopology(network, nodes)

}
