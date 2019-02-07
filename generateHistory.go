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
const txPrefix = "txLogN"
const pingPrefix = "pingLogN"
const timeFormat = "15:04:05.999"

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
	Width int16
	Children int16
}

type PingPacket struct {
	SentTimestamp time.Time
	ReceiveTimestamp time.Time
	Sender uint16
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

func getBlockLines(node int) (val []string) {
	data := readLogFile(node)
	/*s := strings.Split(data, "\n")
	max := 1154482
	if max > len(s) {
		max = len(s)
	}
	if len(s)>1152000 {
		val = make([]string, 0)
		for i:=1152000; i<max; i++ {
			if len(s[i]) > 0 {
				buff := make([]byte, len(s[i]))
				_ = copy(buff, s[i])
				val = append(val, string(buff))
			}
		}
	}*/
	return strings.Split(data, "\n")
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

		if len(entry)>2 {

			if block, contained = blockchain[entry[1]]; !contained {
				block.AcceptDelays = make([]Delay, 0, 240)
				block.DiscoveryDelays = make([]Delay, 0, 240)
				block.Width = -1
				block.Children = 0
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

func createBlockChain(nodeAmount int) map[string]Block {

	blockchain := make(map[string]Block)

	for i:=0;i<nodeAmount; i++ {
		addNodeInfo(blockchain, i)
	}

	computeDelays(blockchain)

	return blockchain
}

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
		fmt.Println(fmt.Sprintf("Found Root with hash %s and parent %s", hash, block.Parent))
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

	list := make([][]string, len(heights))

	for k, v := range heights {
		if list[v] == nil {
			list[v] = []string{k}
		} else {
			list[v] = append(list[v], k)
		}
	}

	return list
}

func createPropagationFile() (outFile *os.File) {
	var err error

	if outFile, err = os.Create(os.Args[2]+"history"); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file %shistory\n %s\n", os.Args[2], err.Error()))
	}

	return outFile
}

func writeToFile(file *os.File, content string) {
	if _, err := file.Write([]byte(content)); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to %s.\n %s\n", file.Name(), err.Error()))
	}
}

func solveWidth(block *Block, blockchain map[string]Block){


	parentBlock, ok := blockchain[block.Parent]

	if ok {
		if parentBlock.Width < 0 {
			solveWidth(&parentBlock, blockchain)
		}
		parentBlock.Children++
		blockchain[block.Parent] = parentBlock
		if parentBlock.Children <= 1 {
			block.Width = parentBlock.Width+1
		} else {
			block.Width = 1
		}
	} else {
		block.Width = 1
	}

	return

}

func updateWidthInfo(blockchain map[string]Block, list [][]string) {

	var i int
	for i=1499; i<len(list); i++ {
		if len(list[i])==1 {
			b := blockchain[list[i][0]]
			b.Width = 0
			blockchain[list[i][0]] = b
			break;
		}
	}

	if i==len(list) {
		fmt.Println("Could not find blockchain tip, assigning randomly")
		if i>1500 {
			b := blockchain[list[1500][0]]
			b.Width = 0
			blockchain[list[1500][0]] = b
			i = 1500
		}else {
			i--
			b := blockchain[list[i][0]]
			b.Width = 0
			blockchain[list[i][0]] = b
		}
	}

	parent := blockchain[list[i][0]].Parent

	var block Block
	var ok bool

	for block, ok = blockchain[parent]; ok; block, ok = blockchain[parent] {
		block.Width = 0
		blockchain[parent] = block
		parent = block.Parent
	}

	var hash string

	for hash, block = range(blockchain) {
		if block.Width < 0 {
			solveWidth(&block, blockchain)
			blockchain[hash] = block
		}
	}

}

func printPropagationTimes(blockchain map[string]Block, list [][]string, nodeAmount int) {

	file := createPropagationFile()

	acceptPercentiles := make([]float64, 15)
	discoveryPercentiles := make([]float64, 15)
	meanPercentilesDisc := make([]float64, 15)
	meanPercentilesAcc := make([]float64, 15)
	amountOfPercentilesDisc := make([]int, 15)
	amountOfPercentilesAcc := make([]int, 15)
	var block Block

	var totWaitTimeAcc uint64 = 0
	var totWaitTimeDisc uint64 = 0
	var sampleSizeAcc uint64 = 0
	var sampleSizeDisc uint64 = 0

	nodeAmount /= 10

	for i:=0; i<len(list) && len(list[i])>0 && i<1501; i++ {

		for j:=0; j<len(list[i]); j++ {

			block, _ = blockchain[list[i][j]]

			writeToFile(file, fmt.Sprintf("%d %s %s %d\n", i, list[i][j], block.Parent, block.NTx))

			writeToFile(file,"Discovery times:")

			for k:=0; k<len(block.DiscoveryDelays)-1; k++ {
				writeToFile(file, fmt.Sprintf(" %d: %f s,", block.DiscoveryDelays[k].node, block.DiscoveryDelays[k].delay.Seconds()))
				if p:= k+1; k>0 && p%nodeAmount==0 {
					discoveryPercentiles[p/nodeAmount] = block.DiscoveryDelays[k].delay.Seconds()
				}
				totWaitTimeDisc += uint64(block.DiscoveryDelays[k].delay.Nanoseconds()/1000000)
				sampleSizeDisc++
			}
			if k:=len(block.DiscoveryDelays)-1; k>=0 {
				writeToFile(file, fmt.Sprintf("%d: %f s\n", block.DiscoveryDelays[k].node, block.DiscoveryDelays[k].delay.Seconds()))
				if p:= k+1; k>0 && p%nodeAmount==0 {
					discoveryPercentiles[p/nodeAmount] = block.DiscoveryDelays[k].delay.Seconds()
				}
				totWaitTimeDisc += uint64(block.DiscoveryDelays[k].delay.Nanoseconds()/1000000)
				sampleSizeDisc++
			}

			sampleSizeDisc--

			writeToFile(file,"Discovery percentiles:")

			for k:=1; k<=10 && nodeAmount*k<=len(block.DiscoveryDelays); k++ {
				writeToFile(file, fmt.Sprintf(" %d: %f s,", k*10, discoveryPercentiles[k]))
				meanPercentilesDisc[k] += discoveryPercentiles[k]
				amountOfPercentilesDisc[k]++
			}

			writeToFile(file,"\n")

			for k:=0; k<len(block.AcceptDelays)-1; k++ {
				writeToFile(file, fmt.Sprintf("%d: %f s, ", block.AcceptDelays[k].node, block.AcceptDelays[k].delay.Seconds()))
				if p:= k+1; k>0 && p%nodeAmount==0 {
					acceptPercentiles[p/nodeAmount] = block.AcceptDelays[k].delay.Seconds()
				}
				totWaitTimeAcc += uint64(block.AcceptDelays[k].delay.Nanoseconds()/1000000)
				sampleSizeAcc++
			}
			if k:=len(block.AcceptDelays)-1; k>=0 {
				writeToFile(file, fmt.Sprintf("%d: %f s\n", block.AcceptDelays[k].node, block.AcceptDelays[k].delay.Seconds()))
				if p:= k+1; k>0 && p%nodeAmount==0 {
					acceptPercentiles[p/nodeAmount] = block.AcceptDelays[k].delay.Seconds()
				}
				totWaitTimeAcc += uint64(block.AcceptDelays[k].delay.Nanoseconds()/1000000)
				sampleSizeAcc++
			}

			sampleSizeAcc--

			writeToFile(file,"Acceptance percentiles: ")

			for k:=1; k<=10 && nodeAmount*k<=len(block.AcceptDelays); k++ {
				writeToFile(file, fmt.Sprintf("%d: %f s, ", k*10, acceptPercentiles[k]))
				meanPercentilesAcc[k] += acceptPercentiles[k]
				amountOfPercentilesAcc[k]++
			}

			writeToFile(file,"\n")

			if len(block.AcceptDelays) > nodeAmount*10 {
				fmt.Println(i, list[i][j])
			}

		}
	}

	totWaitTimeDisc /= sampleSizeDisc

	writeToFile(file,fmt.Sprintf("Mean Block Discovery Time: %d.%ds\n", totWaitTimeDisc/1000, totWaitTimeDisc%1000))

	totWaitTimeAcc /= sampleSizeAcc

	writeToFile(file,fmt.Sprintf("Mean Block Acceptance Time: %d.%ds\n", totWaitTimeAcc/1000, totWaitTimeAcc%1000))

	writeToFile(file,"Mean Discovery percentiles: ")

	for k:=1; k<=10; k++ {
		writeToFile(file, fmt.Sprintf("%d: %f s, ", k*10, meanPercentilesDisc[k]/float64(amountOfPercentilesDisc[k])))
	}

	writeToFile(file,"\n")

	writeToFile(file,"Mean Acceptance percentiles: ")

	for k:=1; k<=10; k++ {
		writeToFile(file, fmt.Sprintf("%d: %f s, ", k*10, meanPercentilesAcc[k]/float64(amountOfPercentilesAcc[k])))
	}

	writeToFile(file,"\n")

}

func createBlockTimeFile() (outFile *os.File) {
	var err error

	if outFile, err = os.Create(os.Args[2]+"blockTimes"); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file %sblockTimes\n %s\n", os.Args[2], err.Error()))
	}

	return outFile
}

func writeBlockTimes(blockchain map[string]Block, list [][]string) {

	updateWidthInfo(blockchain, list)

	forksInfo := make([]int16, 16)

	blockTimesFile := createBlockTimeFile()

	block := blockchain[list[0][0]]

	var tmpBlock Block

	initTime := block.Time

	lastTime := initTime

	var meanDiff int64 = 0

	var i int64
	var j int

	var d int64 = 0

	pendingTimes := make(map[time.Duration]bool)

	promprom := 0.0
	
	avgs := make([]int, 10)

	for i = 0; i<int64(len(list)) && len(list[i])>0 && i<1501; i++ {

		for j=0; j<len(list[i]); j++ {
			tmpBlock = blockchain[list[i][j]]
			if tmpBlock.Width == 0 {
				block = tmpBlock
			} else {
				forksInfo[tmpBlock.Width]++
			}
		}

		for pendingTime, _ := range(pendingTimes) {
			if pendingTime <= block.Time {
				d++
				delete(pendingTimes, pendingTime)
			}
		}

		for j=0; j<len(list[i]); j++ {
			tmpBlock = blockchain[list[i][j]]
			if tmpBlock.Time <= block.Time {
				d++
			} else {
				pendingTimes[tmpBlock.Time] = true
			}
		}

		prom := float64(block.NTx)/74.83

		if i>=30 {
			promprom += prom
			for j:=0; j<10 && float64(j*10)<=prom; j++ {
				avgs[j]++
			}
		}

		s := fmt.Sprintf("%d %s %d %.3f", len(list[i]), time.Unix(0, int64(block.Time)).Format(timeFormat), block.NTx, prom)

		diff := block.Time - initTime

		s += fmt.Sprintf(" %d:%d:%d ", int64(diff.Hours()), int64(diff.Minutes())%60, int64(diff.Seconds()+0.5)%60)

		diff = block.Time - lastTime

		meanDiff += diff.Nanoseconds()

		s += fmt.Sprintf("+%d seconds ", int64(diff.Seconds()+0.5))

		if i>0 {
			s += fmt.Sprintf("- Mean Block Diff: %d seconds, Mean Height Diff: %d seconds\n", ((meanDiff/(d))+500000000)/1000000000, ((meanDiff/(i))+500000000)/1000000000)
		} else {
			s += "\n"
		}

		writeToFile(blockTimesFile, s)

		lastTime = block.Time
	}

	writeToFile(blockTimesFile, fmt.Sprintf("Mean Diff of Heights: %d seconds\nMean Diff of blocks: %d seconds \nMean Percentage of Fullness:%f\n", ((meanDiff/i)+500000000)/1000000000, ((meanDiff/d)+500000000)/1000000000, promprom/float64(i-30)))
	for j:=0; j<10; j++ {
		writeToFile(blockTimesFile, fmt.Sprintf("Percentage of blocks above %d of fullness: %f\n", j*10, (float64(avgs[j])/float64(i-30))*100.0))
	}

	for j:=1; j<len(forksInfo) && forksInfo[j]>0; j++ {
		writeToFile(blockTimesFile, fmt.Sprintf("Number of Forks of height %d: %d\n", j, forksInfo[j]))
	}

}

func createTxFile() (outFile *os.File) {
	var err error

	if outFile, err = os.Create(os.Args[2]+"txIdleTimes"); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file %stxIdleTimes\n %s\n", os.Args[2], err.Error()))
	}

	return outFile
}

func readTxFile(node int) string {
	if b, err := ioutil.ReadFile(fmt.Sprintf("%s%s%d", os.Args[1], txPrefix, node)); err == nil {
		return string(b)
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse log file.\n %s\n", err.Error()))
		return ""
	}
}

func getTxLogLines(node int) (val []string) {
	data := readTxFile(node)
	return strings.Split(data, "\n")
}

func addTxData(file *os.File, node int) {

	lines := getTxLogLines(node)

	if len(lines)>1 {
		lines = strings.Split(lines[1], " ")
		if len(lines)>9 {
			writeToFile(file, lines[9]+"\n")
		}
	}

}

func printTxData(nodeAmount int) {

	file := createTxFile()

	for i:=0; i<nodeAmount; i++ {
		addTxData(file, i)
	}

}

func printBlockchainData(nodeAmount int) {

	blockchain := createBlockChain(nodeAmount)

	list := getHeightList(blockchain)

	writeBlockTimes(blockchain, list)

	printPropagationTimes(blockchain, list, nodeAmount)
}

func readPingLogFile(node int) []byte {
	if b, err := ioutil.ReadFile(fmt.Sprintf("%s%s%d", os.Args[1], pingPrefix, node)); err == nil {
		return b
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse log file.\n %s\n", err.Error()))
		return nil
	}
}

func decodeTimestamp(b []byte) (p int64) {
	p = int64(b[7])
	p |= int64(b[6]) << 8
	p |= int64(b[5]) << 16
	p |= int64(b[4]) << 24
	p |= int64(b[3]) << 32
	p |= int64(b[2]) << 40
	p |= int64(b[1]) << 48
	p |= int64(b[0]) << 56
	return p
}

func decodeSender(b []byte) (p uint16) {
	p = uint16(b[1])
	p |= uint16(b[0]) << 8
	return p
}

func getPingPackets(node int) []PingPacket {
	data := readPingLogFile(node)

	var packet PingPacket
	maxIndex := len(data)-18

	if maxIndex < 0 {maxIndex = 0}

	pingPackets := make([]PingPacket, 0, maxIndex+1)

	for i:=0; i<=maxIndex; i+=18 {
		packet.SentTimestamp = time.Unix(0, decodeTimestamp(data[i:i+8]))
		packet.ReceiveTimestamp = time.Unix(0, decodeTimestamp(data[i+8:i+16]))
		packet.Sender = decodeSender(data[i+16:i+18])

		pingPackets = append(pingPackets, packet)
	}

	return pingPackets
}

func addPingData(node int, pings [][][]int16) {

	packets := getPingPackets(node)

	for i:=0; i<len(packets); i++ {
		pings[int(packets[i].Sender)][node] = append(pings[int(packets[i].Sender)][node], int16((packets[i].ReceiveTimestamp.Sub(packets[i].SentTimestamp).Nanoseconds()+500000)/1000000))
	}

	return
}

func createPingFile() (outFile *os.File) {
	var err error

	if outFile, err = os.Create(os.Args[2]+"pingTimes"); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file %spingTimes\n %s\n", os.Args[2], err.Error()))
	}

	return outFile
}

func printPingData() {

	hostAmount, _ := strconv.Atoi(os.Args[4])

	pings := make([][][]int16, hostAmount)

	for i:=0; i<hostAmount; i++ {
		pings[i] = make([][]int16, hostAmount)
		for j:=0; j<hostAmount; j++ {
			pings[i][j] = make([]int16, 0)
		}
	}

	for node:=0; node<hostAmount; node++ {
		addPingData(node, pings)
	}

	pingsFile := createPingFile()

	var k int
	var j int

	for i:=0; i<hostAmount; i++ {
		for j=0; j<hostAmount; j++ {
			if i!=j {
				writeToFile(pingsFile, fmt.Sprintf("%d a %d:", i, j))
				k = 0
				for l:=len(pings[i][j])-1; k<l; k++ {
					writeToFile(pingsFile, fmt.Sprintf(" %d", pings[i][j][k]))
				}
				if k>0 {
					writeToFile(pingsFile, fmt.Sprintf(" %d\n", pings[i][j][k]))
				}
			}
		}
	}
}

func main(){

	nodeAmount, _ := strconv.Atoi(os.Args[3])

	printBlockchainData(nodeAmount)

	printTxData(nodeAmount)

	printPingData()
}
