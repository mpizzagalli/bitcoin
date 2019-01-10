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
const timeFormat = "15:04:05.999999"

type Delay struct {
	delay time.Duration
	node byte
}

type Block struct {
	AcceptDelays DelaySort
	Time time.Duration
	DiscoveryDelays DelaySort
	NTx int64
	Parent string
}
/*
type Blockchain struct {
	Blocks []Block
	Index	map[string]int
}

func (b *Blockchain) at(h string) Block {
	return b.Blocks[b.Index[h]]
}

func (b *Blockchain) insert(h string) Block {
	return b.Blocks[b.Index[h]]
}*/

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

func getBlockLines(node int) (val []string) {
	data := readLogFile(node)
	s := strings.Split(data, "\n")
	max := 1154482
	if max > len(s) {
		max = len(s)
	}
	if len(s)>1152000 {
		val = make([]string, 0)
		for i:=1152000; i<max; i++ {
			if len(s[i]) > 0 /*&& s[i][0] != '2'*/ {
				buff := make([]byte, len(s[i]))
				_ = copy(buff, s[i])
				val = append(val, string(buff))
			}
		}
	}
	return
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

	for i:=1; i<len(lines); i++ {

		entry = strings.Split(lines[i], " ")

		if /*entry[0] != "2" && */len(entry)>2 {

			if block, contained = blockchain[entry[1]]; !contained {
				block.AcceptDelays = make([]Delay, 0, 240)
				block.DiscoveryDelays = make([]Delay, 0, 240)
			}

			if entry[0]=="0" {
				block.AcceptDelays = append(block.AcceptDelays, Delay{time.Duration(getNsec(entry[4])), byte(node)})
				block.Time = time.Duration(getNsec(entry[4]))
				block.DiscoveryDelays = append(block.DiscoveryDelays, Delay{time.Duration(getNsec(entry[4])), byte(node)})
				block.NTx, _ = strconv.ParseInt(entry[3], 10, 64)
				block.Parent = entry[2]
			} else if entry[0]=="1" {
				block.AcceptDelays = append(block.AcceptDelays, Delay{time.Duration(getNsec(entry[3])), byte(node)})
				block.Parent = entry[2]
			} else {
				block.DiscoveryDelays = append(block.DiscoveryDelays, Delay{time.Duration(getNsec(entry[2])), byte(node)})
			}

			blockchain[entry[1]] = block
		}
	}

}

func computeDelays(blockchain map[string]Block) {

	for _, v := range blockchain {
		for i:=0; i<len(v.AcceptDelays); i++ {
			v.AcceptDelays[i].delay = v.AcceptDelays[i].delay - v.Time
		}
		sort.Sort(v.AcceptDelays)
		for i:=0; i<len(v.DiscoveryDelays); i++ {
			v.DiscoveryDelays[i].delay = v.DiscoveryDelays[i].delay - v.Time
		}
		sort.Sort(v.DiscoveryDelays)
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
/*
func readHieghtFile() string {
	if b, err := ioutil.ReadFile(os.Args[3]); err == nil {
		return string(b)
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse log file.\n %s\n", err.Error()))
		return ""
	}
}*/

func getHeight(blockchain map[string]Block, heights map[string]int, block *Block, hash string) int {

	if h, ok := heights[hash]; ok {
		return h
	}

	if h, ok := heights[block.Parent]; ok {
		heights[hash] = h+1
		return h+1
	}

	if parent, ok := blockchain[block.Parent]; ok {
		p := getHeight(blockchain, heights, &parent, block.Parent)+1
		heights[hash] = p
		return p
	} else {
		fmt.Println(fmt.Sprintf("Found Root with hash %s and parent %s ", hash, block.Parent))
		heights[hash] = 0
		return 0
	}
}

func calculateHeights(blockchain map[string]Block) map[string]int {

	heights := make(map[string]int)

	for k, v := range blockchain {
		getHeight(blockchain, heights, &v, k)
	}

	return heights
}

func getHeightList(blockchain map[string]Block) [][]string {
	//return strings.Split(readHieghtFile(), "\n")

	heights := calculateHeights(blockchain)

	list := make([][]string, 1500)

	for k, v := range heights {
		if list[v] == nil {
			list[v] = make([]string, 0, 1)
		}
		list[v] = append(list[v], k)
	}

	return list
}

func createPropagationFile() (outFile *os.File) {
	var err error

	if outFile, err = os.Create(os.Args[3]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file %s\n %s\n", os.Args[4], err.Error()))
	}

	return outFile
}

func writeToFile(file *os.File, content string) {
	if _, err := file.Write([]byte(content)); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to %s.\n %s\n", file.Name(), err.Error()))
	}
}

func printPropagationTimes(blockchain map[string]Block, list [][]string) {

	file := createPropagationFile()

	//var hashes []string
	var block Block
	//var found bool

	for i:=0; i<len(list) && len(list[i])>0; i++ {

		for j:=0; j<len(list[i]); j++ {

			block, _ = blockchain[list[i][j]]

			//if block, found = blockchain[list[i][j]]; found {
				writeToFile(file, fmt.Sprintf("%d %s %s %d\n", i, list[i][j], block.Parent, block.NTx))

				for k:=0; k<len(block.AcceptDelays)-1; k++ {
					writeToFile(file, fmt.Sprintf("%d: %f s, ", block.AcceptDelays[k].node, block.AcceptDelays[k].delay.Seconds()))
				}
				if k:=len(block.AcceptDelays)-1; k>=0 {
					writeToFile(file, fmt.Sprintf("%d: %f s\n", block.AcceptDelays[k].node, block.AcceptDelays[k].delay.Seconds()))
				}

				for k:=0; k<len(block.DiscoveryDelays)-1; k++ {
					writeToFile(file, fmt.Sprintf("%d: %f s, ", block.DiscoveryDelays[k].node, block.DiscoveryDelays[k].delay.Seconds()))
				}
				if k:=len(block.DiscoveryDelays)-1; k>=0 {
					writeToFile(file, fmt.Sprintf("%d: %f s\n", block.DiscoveryDelays[k].node, block.DiscoveryDelays[k].delay.Seconds()))
				}
			//} else if len(list[i][j]) > 1 {
		//		os.Stderr.WriteString(fmt.Sprintf("Failed to parse block %s.\n", list[i][j]))
		//	}

		}
	}

}

func createBlockTimeFile() (outFile *os.File) {
	var err error

	if outFile, err = os.Create(os.Args[4]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file %s\n %s\n", os.Args[4], err.Error()))
	}

	return outFile
}

func writeBlockTimes(blockchain map[string]Block, list [][]string) {

	blockTimesFile := createBlockTimeFile()

	block := blockchain[list[0][0]]

	var tmpBlock Block

	initTime := block.Time

	lastTime := initTime

	var meanDiff int64 = 0

	var i int64
	var j int
	var max time.Duration

	var d int64 = 0;

	for i = 0; i<int64(len(list)) && len(list[i])>0 && i<1201; i++ {

		block = blockchain[list[i][0]]

		max = block.Time

		for j=1; j<len(list[i]); j++ {
			tmpBlock = blockchain[list[i][j]]
			if tmpBlock.Time>max {
				max = tmpBlock.Time
				block = tmpBlock
			}
		}

		s := fmt.Sprintf("%d %s %d", len(list[i]), time.Unix(0, int64(block.Time)).Format(timeFormat), block.NTx)

		diff := block.Time - initTime

		s += fmt.Sprintf(" %d:%d:%d ", int64(diff.Hours()), int64(diff.Minutes())%60, int64(diff.Seconds()+0.5)%60)

		diff = block.Time - lastTime

		meanDiff += diff.Nanoseconds()

		s += fmt.Sprintf("+%d seconds ", int64(diff.Seconds()+0.5))

		d += int64(len(list[i]))

		if i>0 {
			s += fmt.Sprintf("- Mean Diff: %d seconds\n", ((meanDiff/(d))+500000000)/1000000000)
		} else {
			s += "\n"
		}

		writeToFile(blockTimesFile, s)

		lastTime = block.Time
	}

	//writeToFile(heightFile, fmt.Sprintf("Mean Diff: %d seconds\n", ((meanDiff/i)+500000000)/1000000000))
}

func main(){

	nodeAmount, _ := strconv.ParseInt(os.Args[2], 10, 64)

	blockchain := createBlockChain(os.Args[1], int(nodeAmount))

	list := getHeightList(blockchain)

	writeBlockTimes(blockchain, list)

	printPropagationTimes(blockchain, list)
}
