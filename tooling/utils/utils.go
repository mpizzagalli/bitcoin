package utils

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	Config "../config"
)

var config = Config.GetConfiguration()
var nodeExecutionDir = config.NodeExecutionDir
var addressesDir = config.AddressesDir

// Mixture of things used all over our codebase
// FIXME: I should refactor how this is done someday

// Executes a command returning it's result
func ExecCmd(cmd *exec.Cmd) []byte {
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	stdOut, execErr := cmd.Output()
	if execErr != nil || stdErr.Len() > 0 {
		os.Stderr.WriteString(fmt.Sprintf("Error executing command.\n %s : %s\n", execErr.Error(), stdErr.String()))
	}
	return stdOut
}

// Creates a random number generator
func CreateRng() *rand.Rand {
	// ugly hack to make all nodes use a different seed: get the seed from crypto/rand
	buf := make([]byte, 8)
	_, _ = crand.Read(buf)

	seed := int64(binary.LittleEndian.Uint64(buf))

	return rand.New(rand.NewSource(seed))
}

func CheckError(e error) {
	if e != nil {
		fmt.Println("Something went wrong, error:", e)
		panic(e)
	}
}

func ParseStartTime(timestamp string) time.Time {
	asd, e := strconv.ParseInt(timestamp, 10, 64)
	CheckError(e)
	return time.Unix(asd, 0)
}

type Node struct {
	id        string
	ip        string
	port      string
	rpcport   string
	addresses []string
}

func newNode(idInt int, addresses []string, portInt int) Node {
	id := strconv.Itoa(idInt)
	port := strconv.Itoa(portInt)
	rpcport := strconv.Itoa(portInt + 1)
	n := Node{id: id, ip: "127.0.0.1", port: port, rpcport: rpcport, addresses: addresses}
	return n
}

func GetAddresses(nodeID int) (addresses []string) {

	if addressesBytes, err := ioutil.ReadFile(addressesDir + "/addrN" + strconv.Itoa(nodeID)); err == nil {
		addresses = strings.Split(string(addressesBytes), "\n")
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse address file.\n %s\n", err.Error()))
	}
	// time.Sleep(time.Duration(10-nodeNumber) * time.Second)
	time.Sleep(time.Duration(1) * time.Second)

	return
}

func ParseNodeInfo(nodeID int) Node {
	addr := GetAddresses(nodeID)
	port := 8330 + nodeID*2

	return newNode(nodeID, addr, port)
}

func ParseNodesInfo(totalNodes int) []Node {
	result := make([]Node, totalNodes)
	for i := 0; i < totalNodes; i++ {
		result[i] = ParseNodeInfo(i)
	}
	return result
}

func WriteTraceOut(nodeID string, traceOut *os.File, startTime time.Time) {
	diff := time.Now().Sub(startTime)
	_, err := traceOut.WriteString(strconv.FormatInt(diff.Milliseconds(), 10) + " " + nodeID + "\n")
	CheckError(err)
}

func MineBlock(nodeInfo Node, traceOut *os.File, startTime time.Time) {
	cmd := exec.Command("bash", nodeExecutionDir+"/bitcoindo.sh", nodeInfo.id, "generatetoaddress", "1", nodeInfo.addresses[0])
	// fmt.Println(cmd.String())
	cmd.Run()
	WriteTraceOut(nodeInfo.id, traceOut, startTime)
}

func MineBlock2(nodeInfo Node, fn func(string)) {
	cmd := exec.Command("bash", nodeExecutionDir+"/bitcoindo.sh", nodeInfo.id, "generatetoaddress", "1", nodeInfo.addresses[0])
	// fmt.Println(cmd.String())
	cmd.Run()
	fn(nodeInfo.id)
}

func WriteTraceOutFn(traceFileOutName string, startTime time.Time) func(string) {
	traceOutFile, e := os.Create(traceFileOutName)
	CheckError(e)
	defer traceOutFile.Close()

	return func(id string) {
		diff := time.Now().Sub(startTime)
		_, err := traceOutFile.WriteString(strconv.FormatInt(diff.Milliseconds(), 10) + " " + id + "\n")
		CheckError(err)
	}
}
