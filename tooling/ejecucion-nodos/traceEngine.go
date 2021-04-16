package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	Config "../config"
	Utils "../utils"
)

var config = Config.GetConfiguration()
var nodeExecutionDir = config.NodeExecutionDir
var addressesDir = config.AddressesDir

type node struct {
	id        string
	ip        string
	port      string
	rpcport   string
	addresses []string
}

func newNode(idInt int, addresses []string, portInt int) node {
	id := strconv.Itoa(idInt)
	port := strconv.Itoa(portInt)
	rpcport := strconv.Itoa(portInt + 1)
	n := node{id: id, ip: "127.0.0.1", port: port, rpcport: rpcport, addresses: addresses}
	return n
}

func getAddresses(nodeID int) (addresses []string) {

	if addressesBytes, err := ioutil.ReadFile(addressesDir + "/addrN" + strconv.Itoa(nodeID)); err == nil {
		addresses = strings.Split(string(addressesBytes), "\n")
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse address file.\n %s\n", err.Error()))
	}
	// time.Sleep(time.Duration(10-nodeNumber) * time.Second)

	return
}

func parseNodesInfo(totalNodes int) []node {
	result := make([]node, totalNodes)
	for i := 0; i < totalNodes; i++ {
		addr := getAddresses(i)
		port := 8330 + i*2

		result[i] = newNode(i, addr, port)
	}
	return result
}

func writeTraceOut(nodeID string, traceOut *os.File, startTime time.Time) {
	diff := time.Now().Sub(startTime)
	_, err := traceOut.WriteString(strconv.FormatInt(diff.Milliseconds(), 10) + " " + nodeID + "\n")
	Utils.CheckError(err)
}

func mineBlock(nodeInfo node, traceOut *os.File, startTime time.Time) {
	cmd := exec.Command("bash", nodeExecutionDir+"/bitcoindo.sh", nodeInfo.id, "generatetoaddress", "1", nodeInfo.addresses[0])
	// fmt.Println(cmd.String())
	cmd.Run()
	writeTraceOut(nodeInfo.id, traceOut, startTime)
}

func processTrace(traceIn *os.File, totalNodes int, nodesInfo []node, traceOut *os.File, startTime time.Time) {
	scanner := bufio.NewScanner(traceIn)
	previousDuration, b := time.ParseDuration("0ms")
	Utils.CheckError(b)

	for scanner.Scan() {
		parsedLine := strings.Fields(scanner.Text())
		nodeID, _ := strconv.Atoi(parsedLine[1])
		blockDuration, e := time.ParseDuration(parsedLine[0] + "ms")
		Utils.CheckError(e)
		diffDuration := blockDuration - previousDuration
		fmt.Println("Scanned line, Node:", nodeID, "after", diffDuration)
		if totalNodes <= nodeID {
			fmt.Println("The trace has a nodeID bigger as expected")
			panic("The trace has a nodeID bigger as expected")
		}
		time.Sleep(diffDuration)
		mineBlock(nodesInfo[nodeID], traceOut, startTime)
		previousDuration = blockDuration
	}
}

func parseStartTime(timestamp string) time.Time {
	asd, _ := strconv.ParseInt(timestamp, 10, 64)
	return time.Unix(asd, 0)
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Missing arguments, usage: go run traceEngine.go tracefileIn #nodes tracefileOut timestart")
		panic("Missing arguments")
	}

	traceFileInName := os.Args[1]
	totalNodes, a := strconv.Atoi(os.Args[2])
	Utils.CheckError(a)

	nodesInfo := parseNodesInfo(totalNodes)

	fmt.Println("nodesInfo: ", nodesInfo)

	traceIn, err := os.Open(traceFileInName)
	Utils.CheckError(err)
	defer traceIn.Close()

	startTime := parseStartTime(os.Args[4])
	traceFileOutName := os.Args[3]

	traceOut, e := os.Create(traceFileOutName)
	Utils.CheckError(e)
	defer traceOut.Close()

	processTrace(traceIn, totalNodes, nodesInfo, traceOut, startTime)

	fmt.Println("traceEngine ended succesfully")
}
