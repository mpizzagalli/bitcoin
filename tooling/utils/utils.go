package utils

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
)

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

func CheckIfError(e error, context string) {
	if e != nil {
		fmt.Println("[", context, "] Something went wrong, error:", e)
		panic(e)
	}
}

func CheckError(e error) {
	if e != nil {
		fmt.Println("Something went wrong, error:", e)
		panic(e)
	}
}
