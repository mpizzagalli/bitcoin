package main

import (
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
var blockIntervalInSeconds = config.BlockIntervalInSeconds

func getAddresses(nodeNumber int) (addresses []string) {

	if addressesBytes, err := ioutil.ReadFile(addressesDir + "/addrN" + os.Args[1]); err == nil {
		addresses = strings.Split(string(addressesBytes), "\n")
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse address file.\n %s\n", err.Error()))
	}

	// Nota: Muy molesto para correr pruebas
	// fmt.Println("Sleep:", time.Duration(10-nodeNumber)*time.Second)
	time.Sleep(time.Duration(10-nodeNumber) * time.Second)

	return
}

// Para obtener nuestro numero de nodo
func getNodeNumber() int {
	n, e := strconv.Atoi(os.Args[1])
	if e != nil {
		os.Stderr.WriteString("Missing node number as argument.\n")
	}
	return n
}

func check(e error) {
	if e != nil {
		fmt.Println("Check error", e)
		panic(e)
	}
}

// Mina bloques de acuerdo al hashing power asignado (simuLambda)
func mineBlocks(addresses []string, nodeNumber int, traceFileName string, startTime time.Time) {

	simuLambda, _ := strconv.ParseFloat(os.Args[2], 64)

	var sleepTime float64
	var sleepSeconds time.Duration
	var sleepNanoseconds time.Duration
	var nextBlockTime time.Duration

	rng := Utils.CreateRng()

	timestamp := time.Now().UnixNano()

	f, e := os.Create(traceFileName)

	check(e)
	fmt.Println("trace file Created")
	// defer f.Close()

	for i := 0; ; i ^= 1 {
		// Hardcodeado que el tiempo entre bloques sea propocional a los 10 min
		sleepTime = (rng.ExpFloat64() / simuLambda) * blockIntervalInSeconds
		sleepSeconds = time.Duration(sleepTime)
		sleepNanoseconds = time.Duration((sleepTime - float64(sleepSeconds)) * 1000000000.0)
		nextBlockTime = time.Duration(timestamp) + sleepSeconds*time.Second + sleepNanoseconds

		time.Sleep(nextBlockTime - time.Duration(time.Now().UnixNano()))

		timestamp = time.Now().UnixNano()
		diff := time.Now().Sub(startTime)

		os.Stdout.WriteString(fmt.Sprintf("[testEngine] %d: Mining a block!\n", nodeNumber))
		_ = exec.Command("bash", nodeExecutionDir+"/bitcoindo.sh", strconv.Itoa(nodeNumber), "generatetoaddress", "1", addresses[i]).Run()
		// a := exec.Command("echo", strconv.FormatInt(diff.Milliseconds(), 10), strconv.Itoa(nodeNumber), ">>", traceFileName).Run()
		_, err := f.WriteString(strconv.FormatInt(diff.Milliseconds(), 10) + " " + strconv.Itoa(nodeNumber) + "\n")
		check(err)
		fmt.Println("Finished writeString after mining")
	}
}

func parseStartTime(timestamp string) time.Time {
	asd, _ := strconv.ParseInt(timestamp, 10, 64)
	return time.Unix(asd, 0)
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
	addresses := getAddresses(nodeNumber)
	startTime := parseStartTime(os.Args[4])
	traceFile := os.Args[3]
	os.Stdout.WriteString("[minerEngine] Finished getting addresses, starting the whole process\n")

	mineBlocks(addresses, nodeNumber, traceFile, startTime)
}
