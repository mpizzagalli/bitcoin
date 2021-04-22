package main

import (
	"os"
	"strconv"
	"time"

	Config "../config"
	Utils "../utils"
)

var config = Config.GetConfiguration()
var blockIntervalInSeconds = config.BlockIntervalInSeconds

// Para obtener nuestro numero de nodo
func getNodeNumber() int {
	n, e := strconv.Atoi(os.Args[1])
	Utils.CheckError(e)
	return n
}

func getSimulatedLambda() float64 {
	simuLambda, e := strconv.ParseFloat(os.Args[2], 64)
	Utils.CheckError(e)
	return simuLambda
}

// Mina bloques de acuerdo al hashing power asignado (simuLambda)
// func mineBlocks(nodeInfo Utils.Node, simuLambda float64, traceFileName string, startTime time.Time) {
func mineBlocks(nodeInfo Utils.Node, simuLambda float64, traceWriteFn func(string)) {

	var sleepTime float64
	var sleepSeconds time.Duration
	var sleepNanoseconds time.Duration
	var nextBlockTime time.Duration

	rng := Utils.CreateRng()

	timestamp := time.Now().UnixNano()

	for i := 0; ; i ^= 1 {
		// Hardcodeado que el tiempo entre bloques sea propocional a los 10 min
		sleepTime = (rng.ExpFloat64() / simuLambda) * blockIntervalInSeconds
		sleepSeconds = time.Duration(sleepTime)
		sleepNanoseconds = time.Duration((sleepTime - float64(sleepSeconds)) * 1000000000.0)
		nextBlockTime = time.Duration(timestamp) + sleepSeconds*time.Second + sleepNanoseconds

		time.Sleep(nextBlockTime - time.Duration(time.Now().UnixNano()))

		timestamp = time.Now().UnixNano()

		// Utils.MineBlock(nodeInfo, traceWriteFn)
		Utils.MineBlock2(nodeInfo, traceWriteFn)
	}
}

/*
	Para levantar mas rapidamente:
	- Comentar el tiempo de espera inicial al cargar las addresses
	- Disminuir el tiempo entre bloques de 10min a algo corto
	- Disminuir el tiempo hasta que se empieza a generar la primer tx
*/
// Manda txs al nodo y mina bloques incluyendolas, hasta que se mata al nodo con una señal de sigterm
func main() {
	nodeNumber := getNodeNumber()
	nodeInfo := Utils.ParseNodeInfo(nodeNumber)
	simuLambda := getSimulatedLambda()

	traceFn := func(id string) { return }

	if len(os.Args) > 3 {
		traceFileName := os.Args[3]
		startTime := Utils.ParseStartTime(os.Args[4])
		var file *os.File
		file, traceFn = Utils.WriteTraceOutFn(traceFileName, startTime)
		defer file.Close()
	}

	mineBlocks(nodeInfo, simuLambda, traceFn)
}
