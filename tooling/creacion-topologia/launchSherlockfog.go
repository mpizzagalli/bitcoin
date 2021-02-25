package main

import (
	"os"
	"fmt"
	"os/exec"
	"bytes"
	Config "../config"
)

var config = Config.GetConfiguration()
var sherlockFogDir = config.SherlockfogDir

func main() {
	if len(os.Args) < 3 {
		os.Stderr.WriteString("Missing fog and ips paths as arguments.\n")
		return
	}

	var fogFilePath = os.Args[1]
	var ipsFilePath = os.Args[2]

	launchFog := exec.Command(/*"bash", "-c", */"python3", sherlockFogDir+"/sherlockfog.py", fogFilePath, "--real-host-list="+ipsFilePath, "--cpu-exclusive=False")//, "> sherlockOut")
	var stdErr bytes.Buffer
	launchFog.Stderr = &stdErr
	b, err := launchFog.Output()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to launch sherlock fog.\n%s : %s\n", err.Error(), stdErr.String()))
	}
	fmt.Println(string(b))
	fmt.Println(stdErr.String())
}