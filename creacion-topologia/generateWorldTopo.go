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

// Parseo del archivo de input a un objeto DistributionJson
func parseJson() (topology DistributionJson) {

	if jsonBytes, err := ioutil.ReadFile(os.Args[2]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse physical layer file.\n %s\n", err.Error()))
	} else if err = json.Unmarshal(jsonBytes, &topology); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse physical layer json.\n %s\n", err.Error()))
	}
	return
}

// Parseo de la cantidad de nodos BTC que deseamos tener
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

// Dada una cantidad de nodos, se los distribuyen entre los paises de acuerdo a los pools y la proporsion de nodos por pais
func calculateHostsPerCountry(data *DistributionJson, amountOfNodes int) map[string]int {

	hostsRemaining := amountOfNodes

	// Mapa: Pais -> cantidad de nodos
	hostsPerCountry := make(map[string]int)

	// Utiliza nodos para tener uno por pool por pais (ya ordenados por proporicion de la red)
	for i:=0; i<len(data.Pools); i++ {
		for j:=0; j<len(data.Pools[i].Nodes) && hostsRemaining > 0; j++ {
			hostsPerCountry[data.Pools[i].Nodes[j].Country]++
			hostsRemaining--
		}
	}

	nodesAmnt := float64(amountOfNodes)

	// Guarda en la data la cantidad de nodos que corresponderia a cada pais
	for i:=0; i<len(data.CountryDistribution); i++ {
		data.CountryDistribution[i].BtcNodes = data.CountryDistribution[i].NetworkShare * nodesAmnt
	}

	// Ordena los paises
	sort.Sort(data.CountryDistribution)

	var z int = -1

	// Aumenta la cantidad de nodos de algunos paises, de acuerdo a la proporsion que deberian tener
	for i := 0; i < len(data.CountryDistribution); i++ {
		desiredAmount := int(data.CountryDistribution[i].BtcNodes)
		if (desiredAmount == 0) {z=i}
		// Maximo entre cantidad de nodos x pools y cantidad por proporsion
		effectiveAmount := maximum(hostsPerCountry[data.CountryDistribution[i].Id], desiredAmount)
		// Libera temporalmente la cantidad de nodos por pool
		hostsRemaining += hostsPerCountry[data.CountryDistribution[i].Id]
		effectiveAmount = minimum(effectiveAmount, hostsRemaining)
		hostsRemaining -= effectiveAmount
		data.CountryDistribution[i].BtcNodes -= float64(effectiveAmount)
		hostsPerCountry[data.CountryDistribution[i].Id] = effectiveAmount
	}

	// Distribuye el resto de los nodos disponibles entre los paises a los que le falte
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

	// Si quedan aun mas se distribuyen uniformemente entre los paises
	for hostsRemaining>0 {
		hostsRemaining--
		hostsPerCountry[data.CountryDistribution[z].Id]++
		z--
		z = z % len(data.CountryDistribution)
	}
	/*for k, v := range (hostsPerCountry) {
		fmt.Println(fmt.Sprintf("%s %d", k, v))
	}*/

	return hostsPerCountry
}

// Configuraciones de los nodos router de cada pais, incluyendo su conexion a los nodos de cada uno de los otros paises
func generateNetwork(data *DistributionJson, amountOfNodes int) (hostNetwork network, countryIdtoHostId map[string]int, hostsPerCountry map[string]int) {

	hostsPerCountry = calculateHostsPerCountry(data, amountOfNodes)

	hostNetwork.Connections = make([]NetworkConnection, 0)

	// Mapa de id de pais a id de nodo de routeo asociado + 1
	countryIdtoHostId = make(map[string]int)

	// Mapa de id de pais a mapa de id de pais a bool, para marcar si agregamos una conexion o no
	addedConnections := make(map[string]map[string]bool)

	for i, j := 0, 0; i< len(data.CountryDistribution); i++ {

		nodesInCountry := hostsPerCountry[data.CountryDistribution[i].Id]
		if nodesInCountry>0 {
			routerHostId := j
			j++
			countryIdtoHostId[data.CountryDistribution[i].Id] = j
			for ;nodesInCountry>0; nodesInCountry-- {
				hostNetwork.Connections = append(hostNetwork.Connections, NetworkConnection{routerHostId, j, int(data.CountryDistribution[i].InnerLatency + 0.5)})
				j++
			}
			addedConnections[data.CountryDistribution[i].Id] = make(map[string]bool)
		}
	}

	// Agregamos ejes tal como esta configurado en CountryLatency
	for i:=0; i<len(data.CountryLatency); i++ {

		countryA := data.CountryLatency[i].A
		countryB := data.CountryLatency[i].B
		
		// Agregamos el enlace entre routers si ambos paises tienen hosts y no agregamos el inverso
		if hostsPerCountry[countryA] > 0 && hostsPerCountry[countryB] > 0 && (!addedConnections[countryB][countryA]) {

			hostA := countryIdtoHostId[countryA]-1
			hostB := countryIdtoHostId[countryB]-1

			hostNetwork.Connections = append(hostNetwork.Connections, NetworkConnection{hostA, hostB, int(data.CountryLatency[i].Latency + 0.5)})

			addedConnections[countryA][countryB] = true
		}
	}

	//Cantidad de hosts es cantidad de nodos + cantidad de paises con nodos
	hostNetwork.Hosts = amountOfNodes + len(countryIdtoHostId)

	return
}

func generateNodes(data *DistributionJson, countryIdtoHostId map[string]int, hostsPerCountry map[string]int, amountOfNodes int) []btcNode {

	btcNodesList := make([]btcNode, 0, amountOfNodes)

	hpLeft := 1.0

	// Agrega de un nodo por cada uno de los mineros de los pools
	// Se realiza esto mientras haya nodos disponibles de los configurados
	// @precambio: no se hacian chequeos de amountOfNodes en ninguno de los dos fors, ocasionando que minimo 45 nodos sean necesarios (por los datos reales utilizados)
	for i:=0; i<len(data.Pools) && amountOfNodes>0; i++ {
		for j:=0; j<len(data.Pools[i].Nodes) && amountOfNodes>0; j++ {
			nodeId := len(btcNodesList)
			nodeHp := data.Pools[i].HPShare * data.Pools[i].Nodes[j].PoolShare
			hpLeft -= nodeHp
			nodeHost := countryIdtoHostId[data.Pools[i].Nodes[j].Country]
			hostsPerCountry[data.Pools[i].Nodes[j].Country]--
			countryIdtoHostId[data.Pools[i].Nodes[j].Country]++
			btcNodesList = append(btcNodesList, btcNode{Id:nodeId, HashingPower:nodeHp, Host:nodeHost})
			amountOfNodes--
		}
	}

	// Aunque esta division pueda quedar en indefinido (o negativo) la variable no se usara en ese caso
	meanHpPerNode := hpLeft/float64(amountOfNodes)

	// Agrega los nodos restantes con hp distribuido uniformemente
	for k, v := range hostsPerCountry {
		for ; v>0; v-- {
			nodeId := len(btcNodesList)
			//nodeHp := meanHpPerNode + rand.NormFloat64() * 0.0002
			nodeHost := countryIdtoHostId[k]
			countryIdtoHostId[k]++
			btcNodesList = append(btcNodesList, btcNode{Id:nodeId, HashingPower:meanHpPerNode, Host:nodeHost, ConnectedTo:make([]int, 0)})
		}
	}

	indexOrder := rand.Perm(len(btcNodesList))

	nodeRndGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	rndSlice := make([]int, 0)

	// Agregamos ejes aleatorios entre los nodos
	for j:=1; j<len(btcNodesList); j++ {

		i := indexOrder[j-1]

		rndSlice = append(rndSlice, []int{i,i,i,i,i,i}...)

		i = indexOrder[j]

		r := nodeRndGen.Int63n(int64(len(rndSlice)))
		nodeA := rndSlice[r]
		btcNodesList[i].ConnectedTo = append(btcNodesList[i].ConnectedTo, nodeA)
		btcNodesList[nodeA].ConnectedTo = append(btcNodesList[nodeA].ConnectedTo, i)
		rndSlice = append(rndSlice[:r], rndSlice[r+1:]...)

		r = nodeRndGen.Int63n(int64(len(rndSlice)))
		nodeB := rndSlice[r]
		if nodeA != nodeB {
			btcNodesList[i].ConnectedTo = append(btcNodesList[i].ConnectedTo, nodeB)
			btcNodesList[nodeB].ConnectedTo = append(btcNodesList[nodeB].ConnectedTo, i)
			rndSlice = append(rndSlice[:r], rndSlice[r+1:]...)
		}
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

/*
	Crea topologia:
	- Nodos router (conectados entre ellos como especificado y con los nodos de cada pais) 
	- Nodos BTC con cierto hashing power (conectados entre ellos de forma medio aleatoria)

	Parametros:
	- Nombre del archivo de output
	- Nombre del archivo de input (igual al producido por genJson.cpp)
	- Cantidad de nodos
 */
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
