package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Correr a mano lo siguiente:")
	a := exec.Command(os.Args[1], os.Args[2:]...)
	fmt.Println(a.String())
	// if err := a.Start(); err != nil {
	// 	fmt.Println("Boomb")
	// 	os.Stderr.WriteString(fmt.Sprintf("Error executing command.\n %s\n", err.Error()))
	// }
}
