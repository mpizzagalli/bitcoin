package main

import (
	"strings"
	"time"
	"os/exec"
	"os"
	"strconv"
	"fmt"
)

const flushInterval = 1024

func getCurrentByteCount() int64 {

	var lines []string
	var i, j int
	var b int64 = 0

	out, _ := exec.Command("cat", "/proc/net/dev").Output()

	if lines = strings.Split(string(out), "\n"); len(lines)>4 {
		lines = strings.Split(lines[4], " ")
		for i, j = 0, 0; i<len(lines) && j < 2; i++ {
			if len(lines[i])>0 {
				j++
			}
		}
		if j==2 {
			b, _ = strconv.ParseInt(lines[i-1], 10, 64)
		}
	}
	return b
}

func printBuffer(buff []int64, logFile *os.File) {
	for i:=0; i<len(buff); i++ {
		logFile.Write([]byte(fmt.Sprintf("%d\n", buff[i])))
	}
}

func createFile() (scriptFile *os.File) {
	var err error

	if scriptFile, err = os.Create("bandwidthLog"); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file bandwidthLog\n%s\n", err.Error()))
	}

	return scriptFile
}

func main() {

	var act int64
	buff := make([]int64, 0, flushInterval)
	logFile := createFile()

	prev := getCurrentByteCount()

	nextTime := time.Duration(time.Now().UnixNano()) + time.Second


	for {
		time.Sleep(nextTime - time.Duration(time.Now().UnixNano()))
		nextTime = time.Duration(time.Now().UnixNano()) + time.Second

		act = getCurrentByteCount()

		if buff = append(buff, act-prev); len(buff)>=flushInterval {
			printBuffer(buff, logFile)
			buff = make([]int64, 0, flushInterval)
		}

		prev = act
	}


}
