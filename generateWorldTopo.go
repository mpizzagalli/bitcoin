package main

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"strconv"
	"sort"
	"math/rand"
	"time"
)

type DistributionJson struct {
	CountryDistribution CountryList `json:"country_distribution"`
	CountryLatency []CountryConnection `json:"country_latency"`
	Pools []BtcPool `json:"pool_distribution"`
}

type BtcCountry struct {
	Id string `json:"id"`
	NetworkShare float64 `json:"share"`
	BtcNodes float64
	InnerLatency float64 `json:"inner_latency"`
}

type CountryList []BtcCountry

func (a CountryList) Len() int           { return len(a) }
func (a CountryList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CountryList) Less(i, j int) bool { return a[i].BtcNodes < a[j].BtcNodes }

type CountryConnection struct {
	A string `json:"a"`
	B string `json:"b"`
	Latency float64 `json:"latency_ms"`
}

type BtcPool struct {
	Id string `json:"id"`
	Nodes []PoolNode `json:"nodes"`
	HPShare float64 `json:"hp_share"`
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

func maximum(a int, b int) int {
	if a>b {
		return a
	} else {
		return b
	}
}

func minimum(a int, b int) int {
	if a<b {
		return a
	} else {
		return b
	}
}

func calculateHostsPerCountry(data *DistributionJson, amountOfNodes int) map[string]int {

	hostsRemaining := amountOfNodes

	hostsPerCountry := make(map[string]int)

	for i:=0; i<len(data.Pools); i++ {
		for j:=0; j<len(data.Pools[i].Nodes); j++ {
			hostsPerCountry[data.Pools[i].Nodes[j].Country]++
			hostsRemaining--
		}
	}

	nodesAmnt := float64(amountOfNodes)

	for i:=0; i<len(data.CountryDistribution); i++ {
		data.CountryDistribution[i].BtcNodes = data.CountryDistribution[i].NetworkShare * nodesAmnt
	}

	sort.Sort(data.CountryDistribution)

	var z int = -1

	for i := 0; i < len(data.CountryDistribution); i++ {
		desiredAmount := int(data.CountryDistribution[i].BtcNodes)
		if (desiredAmount == 0) {z=i}
		effectiveAmount := maximum(hostsPerCountry[data.CountryDistribution[i].Id], desiredAmount)
		hostsRemaining += hostsPerCountry[data.CountryDistribution[i].Id]
		effectiveAmount = minimum(effectiveAmount, hostsRemaining)
		hostsRemaining -= effectiveAmount
		data.CountryDistribution[i].BtcNodes -= float64(effectiveAmount)
		hostsPerCountry[data.CountryDistribution[i].Id] = effectiveAmount
	}

	for z>=0 && hostsRemaining>0 {
		if (data.CountryDistribution[z].BtcNodes > 0) {
			data.CountryDistribution[z].BtcNodes -= 1.0
			hostsPerCountry[data.CountryDistribution[z].Id]++
			hostsRemaining--
		}
		z--
	}

	sort.Sort(data.CountryDistribution)

	z=len(data.CountryDistribution)-1

	for hostsRemaining>0 {
		hostsRemaining--
		hostsPerCountry[data.CountryDistribution[z].Id]++
		z--
		z = z % len(data.CountryDistribution)
	}

	return hostsPerCountry
}

func generateNetwork(data *DistributionJson, amountOfNodes int) (hostNetwork network, countryIdtoHostId map[string]int, hostsPerCountry map[string]int) {

	hostsPerCountry = calculateHostsPerCountry(data, amountOfNodes)

	//Cantidad de hosts es cantidad de nodos + cantidad de paises con nodos
	hostNetwork.Hosts = amountOfNodes

	for _, v := range hostsPerCountry {
		if (v>0) {
			hostNetwork.Hosts++
		}
	}

	hostNetwork.Connections = make([]NetworkConnection, 0)

	countryIdtoHostId = make(map[string]int)

	addedConnections := make(map[string]map[string]bool)

	for i, j := 0, 0; i< len(data.CountryDistribution); i++ {

		nodesInCountry := hostsPerCountry[data.CountryDistribution[i].Id]
		if nodesInCountry>0 {
			countryIdtoHostId[data.CountryDistribution[i].Id] = j //[2]int{j, nodesInCountry}
			routerHostId := j + nodesInCountry
			for j < routerHostId {
				hostNetwork.Connections = append(hostNetwork.Connections, NetworkConnection{routerHostId, j, int(data.CountryDistribution[i].InnerLatency + 0.5)})
				j++
			}
			addedConnections[data.CountryDistribution[i].Id] = make(map[string]bool)
			j++
		}
	}



	for i:=0; i<len(data.CountryLatency); i++ {

		countryA := data.CountryLatency[i].A
		countryB := data.CountryLatency[i].B

		hostA := hostsPerCountry[countryA]
		hostB := hostsPerCountry[countryB]

		//Agregamos el enlace si ambos paises tienen hosts y no agregamos el inverso
		if hostA > 0 && hostB > 0 && (!addedConnections[countryB][countryA]) {

			hostA += countryIdtoHostId[countryA] // agregamos indice base
			hostB += countryIdtoHostId[countryB] // agregamos indice base

			hostNetwork.Connections = append(hostNetwork.Connections, NetworkConnection{hostA, hostB, int(data.CountryLatency[i].Latency + 0.5)})

				addedConnections[countryA][countryB] = true
		}
	}

	return
}

func generateNodes(data *DistributionJson, countryIdtoHostId map[string]int, hostsPerCountry map[string]int, amountOfNodes int) []btcNode {

	btcNodesList := make([]btcNode, 0, amountOfNodes)

	hpLeft := 1.0

	for i:=0; i<len(data.Pools); i++ {
		for j:=0; j<len(data.Pools[i].Nodes); j++ {
			nodeId := len(btcNodesList)
			nodeHp := data.Pools[i].HPShare * data.Pools[i].Nodes[j].PoolShare
			hpLeft -= nodeHp
			nodeHost := countryIdtoHostId[data.Pools[i].Nodes[j].Country]
			hostsPerCountry[data.Pools[i].Nodes[j].Country]--
			countryIdtoHostId[data.Pools[i].Nodes[j].Country]++
			btcNodesList = append(btcNodesList, btcNode{Id:nodeId, HashingPower:nodeHp, Host:nodeHost})
		}
		amountOfNodes -= len(data.Pools[i].Nodes)
	}
	
	meanHpPerNode := hpLeft/float64(amountOfNodes)

	for k, v := range hostsPerCountry {
		for ; v>0; v-- {
			nodeId := len(btcNodesList)
			nodeHp := meanHpPerNode + rand.NormFloat64() * 0.0002
			nodeHost := countryIdtoHostId[k]
			countryIdtoHostId[k]++
			btcNodesList = append(btcNodesList, btcNode{Id:nodeId, HashingPower:nodeHp, Host:nodeHost, ConnectedTo:make([]int, 0)})
		}
	}

	indexOrder := rand.Perm(len(btcNodesList))

	rndSlice := make([]int, 6)

	for i:=0; i<6; i++ {
		rndSlice[i] = indexOrder[0]
	}

	nodeRndGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	for j:=1; j<len(btcNodesList); j++ {

		i := indexOrder[j]

		r := nodeRndGen.Int63n(int64(len(rndSlice)))
		nodeA := rndSlice[r]
		btcNodesList[i].ConnectedTo = append(btcNodesList[i].ConnectedTo, nodeA)
		btcNodesList[nodeA].ConnectedTo = append(btcNodesList[nodeA].ConnectedTo, i)
		rndSlice = append(rndSlice[:r], rndSlice[r+1:]...)

		r = nodeRndGen.Int63n(int64(len(rndSlice)))
		nodeB := rndSlice[r]
		if (nodeA != nodeB) {
			btcNodesList[i].ConnectedTo = append(btcNodesList[i].ConnectedTo, nodeB)
			btcNodesList[nodeB].ConnectedTo = append(btcNodesList[nodeB].ConnectedTo, i)
			rndSlice = append(rndSlice[:r], rndSlice[r+1:]...)
		}

		rndSlice = append(rndSlice, []int{i,i,i,i,i,i}...)
	}

	return btcNodesList
}

func writeTopology(hostNetwork network, nodes []btcNode) {

	jsonData := GraphJson{Network:hostNetwork, BtcNodes:nodes}

	jsonBytes, _ := json.Marshal(jsonData)

	jsonFile := createFile()

	if _, err := jsonFile.Write(append(jsonBytes, '\n')); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to %s.\n %s\n", jsonFile.Name(), err.Error()))
	}
}

func main(){

	if len(os.Args) < 4 {
		os.Stderr.WriteString("Missing Json filenames and total amount of nodes as arguments.\n")
		return
	}

	data := parseJson()

	amountOfNodes := parseAmountOfBtcNodes()

	network, countryIdtoHostId, hostsPerCountry := generateNetwork(&data, amountOfNodes)

	nodes := generateNodes(&data, countryIdtoHostId, hostsPerCountry, amountOfNodes)

	writeTopology(network, nodes)

}
