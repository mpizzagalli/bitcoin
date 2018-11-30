package main

import (
		"strconv"
		"os"
		"os/exec"
		"io/ioutil"
		"strings"
		"fmt"
		"bytes"
)

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

func execCmd(cmd *exec.Cmd) []byte {
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	stdOut, execErr := cmd.Output()
	if execErr != nil || stdErr.Len() > 0 {
		os.Stderr.WriteString(fmt.Sprintf("Error executing command.\n %s : %s\n", execErr.Error(), stdErr.String()))
	}
	return  stdOut
}

func main() {

	n := getAmountOfNodes()
	addresses := getAddresses(n)
	for j:=0; j<1100; j++ {
		for i := 0; i < n; i++ {
			_ = execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "generatetoaddress", "1", addresses[i][0]))
			_ = execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "generatetoaddress", "1", addresses[i][1]))
		}
	}

}
