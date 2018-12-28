package main

import (
	"os/exec"
	"os"
	"fmt"
	"bytes"
	"encoding/json"
	"time"
)

type BlockchainInfo struct {
	Blocks int64 `json:"blocks"`
	Headers int64 `json:"headers"`
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

func haveToWait() bool {
	var stdOut []byte
	var info BlockchainInfo

	stdOut = execCmd(exec.Command("bash", "/home/mgeier/ndecarli/bitcoindo.sh", os.Args[1], "getblockchaininfo"))

	if err := json.Unmarshal(stdOut, &info); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Error unmarshaling signedtransaction json.\n %s\n", err.Error()))
	}

	return info.Blocks != info.Headers
}

func main() {

	if len(os.Args) < 2 {
		os.Stderr.WriteString("Missing node number as argument.\n")
		return
	}

	for haveToWait() {
		time.Sleep(time.Second * 40)
	}

}
