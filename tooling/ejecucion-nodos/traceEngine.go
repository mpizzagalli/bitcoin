package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	Utils "../utils"
)

func processTrace(traceIn *os.File, totalNodes int, nodesInfo []Node, traceOut *os.File, startTime time.Time) {
	scanner := bufio.NewScanner(traceIn)
	previousDuration, b := time.ParseDuration("0ms")
	Utils.CheckError(b)

	for scanner.Scan() {
		parsedLine := strings.Fields(scanner.Text())
		nodeID, _ := strconv.Atoi(parsedLine[1])
		blockDuration, e := time.ParseDuration(parsedLine[0] + "ms")
		Utils.CheckError(e)
		diffDuration := blockDuration - previousDuration
		if totalNodes <= nodeID {
			fmt.Println("The trace has a nodeID bigger as expected")
			panic("The trace has a nodeID bigger as expected")
		}
		time.Sleep(diffDuration)
		Utils.MineBlock(nodesInfo[nodeID], traceOut, startTime)
		previousDuration = blockDuration
	}
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Missing arguments, usage: go run traceEngine.go tracefileIn #nodes tracefileOut timestart")
		panic("Missing arguments")
	}

	traceFileInName := os.Args[1]
	totalNodes, a := strconv.Atoi(os.Args[2])
	Utils.CheckError(a)

	nodesInfo := Utils.ParseNodesInfo(totalNodes)

	traceIn, err := os.Open(traceFileInName)
	Utils.CheckError(err)
	defer traceIn.Close()

	startTime := Utils.ParseStartTime(os.Args[4])
	traceFileOutName := os.Args[3]

	traceOut, e := os.Create(traceFileOutName)
	Utils.CheckError(e)
	defer traceOut.Close()

	processTrace(traceIn, totalNodes, nodesInfo, traceOut, startTime)

	fmt.Println("traceEngine ended succesfully")
}
