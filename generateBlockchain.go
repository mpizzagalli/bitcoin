package main

import (
		"strconv"
		"os"
		"os/exec"
		"io/ioutil"
		"strings"
		"fmt"
		"bytes"
		"time"
		//"encoding/json"
)

type Block struct {
	Tx []string `json:"tx"`
}

type Tx struct {
	Hex string `json:"hex"`
}

func getAmountOfNodes() int {
	n, e := strconv.Atoi(os.Args[1])
	if e != nil {
		os.Stderr.WriteString("Missing node number as argument.\n")
	}
	return n
}

func getAddresses(n int) (addresses [][]string) {
	for i:=0; i<n; i++ {
		if addressesBytes, err := ioutil.ReadFile(fmt.Sprintf("/home/mgeier/ndecarli/addrN%d", i)); err == nil {
			addresses = append(addresses, strings.Split(string(addressesBytes), "\n"))
		} else {
			os.Stderr.WriteString(fmt.Sprintf("Failed to parse addresses file.\n %s\n", err.Error()))
		}
	}

	return
}

func execCmd(cmd *exec.Cmd) {
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	if execErr := cmd.Run(); execErr != nil || stdErr.Len() > 0 {
		os.Stderr.WriteString(fmt.Sprintf("Error executing command %s.\n%s\n%s\n", cmd.Args[2], execErr.Error(), stdErr.String()))
	}
	return
}
/*
const txInputTemplate = "[{\"txid\":\"%s\",\"vout\":0}]"
const txOutputTemplate = "[{\"%s\":9.9999980}, {\"%s\":9.9999980}, {\"%s\":10.0}, {\"%s\":10.0}, {\"%s\":10.0}]"

func txInput(tx string) string {
	return fmt.Sprintf(txInputTemplate, tx)
}

func txOutput(addresses [][]string, node int) string {
	mod := node%2
	base := node - mod
	nodo2 := (base+1)%len(addresses)
	nodo3 := (base+2)%len(addresses)
	nodo4 := (base+3)%len(addresses)
	nodo5 := (base+4)%len(addresses)

	return fmt.Sprintf(txOutputTemplate, addresses[base][mod], addresses[nodo2][mod], addresses[nodo3][mod], addresses[nodo4][mod], addresses[nodo5][mod])
}

func multiplyTransactions(stdOut []byte, node int, addresses [][]string) {

	if len(stdOut) == 0 {
		return
	}

	blockList := make([]string, 0)

	var err error

	if err = json.Unmarshal(stdOut, &blockList); err != nil {
		os.Stderr.WriteString("Could not unmarshal generated blocks list.\n")
	}

	var block Block
	var tx Tx

	for i:=0; i<len(blockList); i++ {

		stdOut = execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "getblock", blockList[i]))

		if err = json.Unmarshal(stdOut, &block); err != nil || len(block.Tx)==0 {
			os.Stderr.WriteString("Could not unmarshal block\n")
		}

		stdOut = execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindoc.sh", os.Args[1], "createrawtransaction", txInput(block.Tx[0]), txOutput(addresses, node)))

		stdOut = execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "signrawtransactionwithwallet", string(stdOut)))

		if err = json.Unmarshal(stdOut, &tx); err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Error unmarshaling signedtransaction json.\n %s\n", err.Error()))
		}

		if err = exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "sendrawtransaction", tx.Hex).Run(); err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Error sending transaction: %s\n", err.Error()))
		}

	}


}*/

func main() {

	n := getAmountOfNodes()
	addresses := getAddresses(n)
	/*var output1 []byte
	var output2 []byte
	var output3 []byte
	var output4 []byte*/
	for j:=0; j<240; j++ {
		for i := 0; i < n; i++ {
			execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "generatetoaddress", "5", addresses[i][0]))
			time.Sleep(10 * time.Millisecond)
			execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "generatetoaddress", "5", addresses[i][1]))
			time.Sleep(10 * time.Millisecond)
			/*multiplyTransactions(output3, i, addresses)
			multiplyTransactions(output4, i, addresses)
			output3 = output1
			output4 = output2*/
		}
	}

}
