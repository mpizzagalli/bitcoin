package main

import (
	"time"
	"os"
	"strconv"
	"io/ioutil"
	"fmt"
	"strings"
	"sort"
)

const filePrefix = "btcCoreLogN"

type Delay struct {
	delay time.Duration
	node byte
}

type Block struct {
	Parent string
	Time time.Time
	Delays DelaySort
	NTx int64
}

type DelaySort []Delay

func (a DelaySort) Len() int           { return len(a) }
func (a DelaySort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DelaySort) Less(i, j int) bool { return a[i].delay < a[j].delay }

func readLogFile(node int) string {
	if b, err := ioutil.ReadFile(fmt.Sprintf("%s%s%d", os.Args[1], filePrefix, node)); err == nil {
		return string(b)
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse log file.\n %s\n", err.Error()))
		return ""
	}
}

func getBlockLines(node int) []string {
	data := readLogFile(node)
	s := strings.Split(data, "\n")
	if (len(s)>1152001) {
		return s[1152001:1154452]
	} else {
		return nil
	}

}

func getNsec(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func addNodeInfo(blockchain map[string]Block, node int){

	lines := getBlockLines(node)
	
	var block Block
	var contained bool
	var entry []string

	for i:=1; i<len(lines); i++{

		entry = strings.Split(lines[i], " ")

		if entry[0] != "2" && len(entry)>2 {

			if block, contained = blockchain[entry[1]]; !contained {
				block.Delays = make([]Delay, 0, 240)
				block.Parent = entry[2]
			}

			if entry[0]=="0" {
				block.Delays = append(block.Delays, Delay{time.Duration(time.Unix(0, getNsec(entry[4])).UnixNano()), byte(node)})
				block.Time = time.Unix(0, getNsec(entry[4]))
				block.NTx, _ = strconv.ParseInt(entry[3], 10, 64)
			} else {
				block.Delays = append(block.Delays, Delay{time.Duration(time.Unix(0, getNsec(entry[3])).UnixNano()), byte(node)})
			}

			blockchain[entry[1]] = block
		}
	}

}

func computeDelays(blockchain map[string]Block) {

	var t time.Time

	for _, v := range(blockchain) {
		for i:=0; i<len(v.Delays); i++ {
			t = time.Unix(0, int64(v.Delays[i].delay))
			v.Delays[i].delay = t.Sub(v.Time)
		}
		sort.Sort(v.Delays)
	}

}

func createBlockChain(baseDir string, nodeAmount int) map[string]Block {

	blockchain := make(map[string]Block)

	for i:=0;i<nodeAmount; i++ {
		addNodeInfo(blockchain, i)
	}

	computeDelays(blockchain)

	return blockchain
}

func printBlockchain(blockchain map[string]Block) {
	fmt.Println(blockchain)
}

func main(){

	nodeAmount, _ := strconv.ParseInt(os.Args[2], 10, 64)

	blockchain := createBlockChain(os.Args[1], int(nodeAmount))

	printBlockchain(blockchain)

}
