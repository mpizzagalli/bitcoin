package utils

import (
	"os"
	"fmt"
	"os/exec"
	"bytes"
	"math/rand"
	crand "crypto/rand"
	"encoding/binary"
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