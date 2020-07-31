package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const bitcoind string = "./src/bitcoind"
const regtest string = "-regtest"
const daemon string = "-daemon"
const pass string = "-rpcpassword=b"
const user string = "-rpcuser=a"
const baseDifficulty string = "-dificulta="
const baseLambda string = "-simuLambda="
const baseDirRoot string = "/Users/mpizzagali/Tesis/regtestData/"
const basePort int = 8330

var mine bool = true
var baseDir string
var currentTimestamp string

func dificulta(cantidadDeZerosRequeridos int) string {
	return baseDifficulty + strconv.Itoa(cantidadDeZerosRequeridos)
}

func simuLambda(lambda float64) string {
	return baseLambda + strconv.FormatFloat(lambda, 'f', 3, 64)
}

func nodeBasePort(nodeNumber int) int {
	return basePort + nodeNumber*2
}

func listenPort(nodeNumber int) string {
	return "-port=" + strconv.Itoa(nodeBasePort(nodeNumber))
}

func rpcPort(nodeNumber int) string {
	return "-rpcport=" + strconv.Itoa(nodeBasePort(nodeNumber)+1)
}

func generateBaseDirFolder() {
	currentTimestamp = time.Now().Format("20060102-15:04:05")
	baseDir = baseDirRoot + currentTimestamp
	mkdir := exec.Command("mkdir", baseDir)
	fmt.Println(mkdir.String())
	if err := mkdir.Run(); err != nil {
		log.Printf("Making directory "+baseDir+" finished with error: %v", err)
	}
	baseDir += "/node"
}

func nodeDataDir(nodeNumber int) string {
	return "-datadir=" + nodeBaseDir(nodeNumber)
}

func nodeBaseDir(nodeNumber int) string {
	return baseDir + strconv.Itoa(nodeNumber)
}

func connectToPeers(peerNodes []int) []string {
	connectionCommands := make([]string, 0, len(peerNodes))
	for i := 0; i < len(peerNodes); i++ {
		connectionCommands = append(connectionCommands, "-addnode=127.0.0.1:"+strconv.Itoa(nodeBasePort(peerNodes[i])))
	}
	return connectionCommands
}

func initCommands(nodeNumber int, peerNodes []int) []string {
	commands := []string{regtest, daemon, pass, user, listenPort(nodeNumber), rpcPort(nodeNumber), nodeDataDir(nodeNumber)}
	//if nodeNumber % 2 == 0 {
	commands = append(commands, dificulta(0))
	//} else{*/
	//commands = append(commands, simuLambda(0.1))
	//}
	return append(commands, connectToPeers(peerNodes)...)
}

func instantiateNode(nodeNumber int, peerNodes []int) {
	mkdir := exec.Command("mkdir", nodeBaseDir(nodeNumber))
	fmt.Println(mkdir.String())
	_ = mkdir.Run()
	cmd := exec.Command(bitcoind, initCommands(nodeNumber, peerNodes)...)
	fmt.Println(cmd.String())
	if err := cmd.Run(); err != nil {
		log.Printf("Instantiation of node "+strconv.Itoa(nodeNumber)+" finished with error: %v", err)
	}
}

func runCommandsOnNode(nodeNumber int, userCommands ...string) {
	baseCommands := []string{regtest, pass, user, rpcPort(nodeNumber)}
	commands := append(baseCommands, userCommands...)
	cmd := exec.Command("./src/bitcoin-cli", commands...)
	fmt.Println(cmd.String())
	var out bytes.Buffer
	cmd.Stdout = &out
	log.Printf("Running ./src/bitcoin-cli " + strings.Join(commands, " "))
	if err := cmd.Run(); err != nil {
		log.Printf("Command on node "+strconv.Itoa(nodeNumber)+" finished with error: %v", err)
	} else {
		fmt.Printf(out.String())
	}
}

func generateBlocks(nodeNumber int, blocksAmmount int) {
	runCommandsOnNode(nodeNumber, "generate", strconv.Itoa(blocksAmmount))
}

func startPoisson(nodeNumber int) {
	rnd := rand.Intn(10)
	time.Sleep(time.Duration(10+rnd) * time.Second)
	for mine {
		rnd := rand.Intn(10)
		generateBlocks(nodeNumber, 1)
		time.Sleep(time.Duration(rnd) * time.Second)
	}
}

func initNode(nodeNumber int, peerNodes []int) {
	instantiateNode(nodeNumber, peerNodes)
	go startPoisson(nodeNumber)
}

func main() {

	generateBaseDirFolder()

	initNode(0, []int{1, 2})
	initNode(1, []int{0, 2})
	initNode(2, []int{0, 1})
	//allPreviousNodes := make([]int, 0, 5)
	//initNode(1, []int{0})
	// for i := 1; i < 5; i++ {
	// 	initNode(i, []int{0, i - 1, i + 1})
	// 	//allPreviousNodes = append(allPreviousNodes, i)
	// }

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Now listening to commands... \n")
	for mine {
		text, _ := reader.ReadString('\n')
		if len(text) > 0 && text[0] == 'q' {
			mine = false
			break
		}
		if commands := strings.Split(text, " "); len(commands) > 0 {
			if nodeNumber, err := strconv.Atoi(commands[0]); err != nil {
				runCommandsOnNode(nodeNumber, commands[1:]...)
			}
		}
	}
}

func main2() {
	fmt.Printf("Printf\n")
	log.Printf("Log Printf\n")
}
