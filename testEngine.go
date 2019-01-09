package main

import (
	"os"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"strings"
	"os/exec"
	"strconv"
	"time"
	"math/rand"
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
)

const txFee float64 = 0.000002
const txSleepMaxAggregate = int64(time.Millisecond * 50)
const txSleepMinimum = time.Millisecond * 2400

type UnspentOutput struct {
	Address string `json:"address"`
	Credit
}

type Credit struct {
	TxId string `json:"txid"`
	Vout int `json:"vout"`
	Amount float64 `json:"amount"`
}

type Tx struct {
	Hex string `json:"hex"`
}

func getAddresses() (addresses []string) {

	if addressesBytes, err := ioutil.ReadFile("/home/mgeier/ndecarli/addrN" + os.Args[1]); err == nil {
		addresses = strings.Split(string(addressesBytes), "\n")
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse address file.\n %s\n", err.Error()))
	}

	return
}

func getNodeNumber() int {
	n, e := strconv.Atoi(os.Args[1])
	if e != nil {
		os.Stderr.WriteString("Missing node number as argument.\n")
	}
	return n
}

func createRng() *rand.Rand {
	// ugly hack to make all nodes use a different seed: get the seed from crypto/rand
	buf := make([]byte, 8)
	_, _ = crand.Read(buf)

	seed := int64(binary.LittleEndian.Uint64(buf))

	return rand.New(rand.NewSource(seed))
}

func mineBlocks(addresses []string) {

	nodeNumber := os.Args[1]
	simuLambda, _ := strconv.ParseFloat(os.Args[2], 64)

	var sleepTime float64
	var sleepSeconds time.Duration
	var sleepNanoseconds time.Duration

	rng := createRng()

	for i := 0;;i ^= 1 {
		sleepTime = (rng.ExpFloat64() / simuLambda)*150.0
		sleepSeconds = time.Duration(sleepTime)
		sleepNanoseconds = time.Duration((sleepTime-float64(sleepSeconds))*1000000000.0)
		time.Sleep(sleepSeconds * time.Second + sleepNanoseconds)
		_ = exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", nodeNumber, "generatetoaddress", "1", addresses[i]).Start()
	}
}

const txInputTemplate = "[{\"txid\":\"%s\",\"vout\":%d}]"

func txInput(credit *Credit) string {
	return fmt.Sprintf(txInputTemplate, credit.TxId, credit.Vout)
}

func txOutput(template string, txCredit float64) string {
	return fmt.Sprintf(template, txCredit-txFee)
}

func outputTemplate(address string) string {
	return "[{\""+address+"\":%f}]"
}

func execCmd(cmd *exec.Cmd) []byte {
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	stdOut, execErr := cmd.Output()
	if execErr != nil || stdErr.Len() > 0 {
		os.Stderr.WriteString(fmt.Sprintf("Error executing command.\n %s : %s\n", execErr.Error(), stdErr.String()))
	}
	return  stdOut
}

func generateTxs(addresses []string) {
	i := 0
	var j [2]int = [2]int{0, 0}
	templates := [2]string{outputTemplate(addresses[1]), outputTemplate(addresses[0])} // posiciones invertidas porque cada direccion le da plata a la otra
	var creditToUse *Credit
	var stdOut []byte //sendrawtransaction
	var timestamp int64
	var tx Tx
	var nextTx time.Duration
	sleepRndGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	nodeNumber := os.Args[1]
	var err error
	unspentOutputs  := make([][]Credit, 1)

	for  {

		if j[i] >= len(unspentOutputs[i]) {
			j[0] = 0
			j[1] = 0
			unspentOutputs = getCredit(addresses)
			for len(unspentOutputs[1]) == 0 || len(unspentOutputs[0]) == 0 {
				time.Sleep(txSleepMinimum)
				unspentOutputs = getCredit(addresses)
			}
		}

		timestamp = time.Now().UnixNano()

		creditToUse = &unspentOutputs[i][j[i]]

		stdOut = execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", nodeNumber, "createrawtransaction", txInput(creditToUse), txOutput(templates[i], creditToUse.Amount)))

		stdOut = execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", nodeNumber, "signrawtransactionwithwallet", string(stdOut)))

		if err = json.Unmarshal(stdOut, &tx); err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Error unmarshaling signedtransaction json.\n %s\n", err.Error()))
		}

		_ = execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", nodeNumber, "sendrawtransaction", tx.Hex))

		j[i]++
		
		i ^= 1

		nextTx = time.Duration(timestamp) + time.Duration(sleepRndGen.Int63n(txSleepMaxAggregate)) + txSleepMinimum

		time.Sleep(nextTx - time.Duration(time.Now().UnixNano()))
	}
}

func getCredit(addresses []string) (credits [][]Credit) {

	btcCmd := exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "listunspent")

	var outputs []UnspentOutput

	if stdOut, err := btcCmd.Output(); err == nil {
		if err = json.Unmarshal(stdOut, &outputs); err != nil {
			os.Stderr.WriteString("Could not unmarshal list of unspent outputs.\n")
		}
	} else {
		os.Stderr.WriteString("Could not retrieve list of unspent outputs.\n")
	}

	credits = make([][]Credit, 2)

	for i:=0; i<len(outputs); i++ {
		if outputs[i].Amount - txFee >= 0.000001 {
			if outputs[i].Address == addresses[0] {
				credits[0] = append(credits[0], outputs[i].Credit)
			} else if outputs[i].Address == addresses[1] {
				credits[1] = append(credits[1], outputs[i].Credit)
			}
		}
	}

	return credits
}

func main(){

	addresses := getAddresses()

	go mineBlocks(addresses)

	generateTxs(addresses)
}
