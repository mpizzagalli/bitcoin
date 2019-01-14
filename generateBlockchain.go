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

func execCmd(cmd *exec.Cmd) {
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	execErr := cmd.Run()
	if execErr != nil || stdErr.Len() > 0 {
		os.Stderr.WriteString(fmt.Sprintf("Error executing command.\n%s\n%s\n", execErr.Error(), stdErr.String()))
	}
	return
}

func main() {

	n := getAmountOfNodes()
	addresses := getAddresses(n)
	for j:=0; j<240; j++ {
		for i := 0; i < n; i++ {
			execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "generatetoaddress", "5", addresses[i][0]))
			time.Sleep(10 * time.Millisecond)
			execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "generatetoaddress", "5", addresses[i][1]))
			time.Sleep(10 * time.Millisecond)
		}
	}

}
