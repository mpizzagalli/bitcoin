package main

import (
	"io/ioutil"
	"os"
	"time"
	"strconv"
	"os/exec"
	"strings"
)

type KindOfEvent byte

const (
	Block KindOfEvent = 0
	Tx KindOfEvent = 1
)

type NodeEvent struct {
	EventMoment time.Time
	Type KindOfEvent
}

func getNewAddr() string {
	addrCmd := exec.Command("/home/mgeier/ndecarli/bitcoindo.sh", os.Args[2], "getnewaddress")
	stdOut, _ := addrCmd.Output()
	sOut := string(stdOut)
	return sOut[0:strings.IndexRune(sOut, '\n')]
}

func parseQueue() (q []NodeEvent){

	startTime := parseStartTime()

	b, _ := ioutil.ReadFile(os.Args[1])
	var tmp NodeEvent
	var offset time.Duration
	j := 0
	s := string(b)
	q = make([]NodeEvent, 0, len(b)/8)

	for i:=0; i<len(s); i=j+1 {
		for j=i; s[j] != '\n'; j++{}
		tmp.Type = KindOfEvent(s[i]-'0')
		offset, _ = time.ParseDuration(s[i+2:j])
		tmp.EventMoment = startTime.Add(offset)
		q = append(q, tmp)
	}

	return q
}

func parseStartTime() time.Time {
	t, _ := strconv.ParseInt(os.Args[3], 10, 64)
	return time.Unix(t, 0)
}

func generateCmdArray() (array [2]*exec.Cmd) {

	walletAddress := getNewAddr()

	array[Block] = exec.Command("/home/mgeier/ndecarli/bitcoindo.sh", os.Args[2], "generatetoaddress", "1", walletAddress)
	array[Tx] = exec.Command("/home/mgeier/ndecarli/bitcoindo.sh", os.Args[2], "sendtoaddress", walletAddress, "1")

	return array
}

func main() {

	events := parseQueue()

	actions := generateCmdArray()

	for i:=0; i<len(events); i++ {
		time.Sleep(events[i].EventMoment.Sub(time.Now()))
		_ = actions[events[i].Type].Run()
	}
}
