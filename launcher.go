package main

import (
	"os"
	"os/exec"
	"fmt"
)

func main(){

	if err := exec.Command(os.Args[1], os.Args[2:]...).Start(); err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Error executing command.\n %s\n", err.Error()))
	}
}
