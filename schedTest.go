package main

import (
	"io/ioutil"
	"os"
	"time"
	"strconv"
	"os/exec"
	"strings"
	"fmt"
)

type KindOfEvent byte

const (
	Block KindOfEvent = 0
	Tx KindOfEvent = 1
)

type NodeEvent struct {
	Offset time.Duration
	Type KindOfEvent
}

func getNewAddr() string {
	addrCmd := exec.Command("/home/mgeier/ndecarli/bitcoinDo.sh", os.Args[2], "getnewaddress")
	stdOut, _ := addrCmd.Output()
	sOut := string(stdOut)
	return sOut[0:strings.IndexRune(sOut, '\n')]
}

func parseQueue() (q []NodeEvent){

	b, _ := ioutil.ReadFile(os.Args[1])
	var tmp NodeEvent
	j := 0
	s := string(b)
	q = make([]NodeEvent, 0, len(b)/8)

	for i:=0; i<len(s); i=j+1 {
		for j=i; s[j] != '\n'; j++{}
		tmp.Type = KindOfEvent(s[i]-'0')
		tmp.Offset, _ = time.ParseDuration(s[i+2:j])
		q = append(q, tmp)
	}

	return q
}

func parseStartTime() time.Time {
	t, _ := strconv.ParseInt(os.Args[3], 10, 64)
	return time.Unix(t, 0)
}


func main() {

	events := parseQueue()

	genBlock := exec.Command("/home/mgeier/ndecarli/bitcoindo.sh", os.Args[2], "generate", "1")
	addTx := exec.Command("/home/mgeier/ndecarli/bitcoindo.sh", os.Args[2], "sendToAddress", getNewAddr(), "1")

	startTime := parseStartTime()
	var nxtEvent time.Time

	for i:=0; i<len(events); i++ {
		nxtEvent = startTime.Add(events[i].Offset)
		time.Sleep(time.Now().Sub(nxtEvent))
		if events[i].Type == Block {
			_ = genBlock.Run()
		} else {
			_ = addTx.Run()
		}
	}
}
