package main

import (
	"os"
	"time"
	"io/ioutil"
	"fmt"
	"strings"
	"strconv"
)

const timeFormat = "15:04:05.999999"

type block struct {
	Hash string
	Parent string
	Time time.Time
}

func createFile(i int) (blkchnFile *os.File) {
	var err error

	if blkchnFile, err = os.Create(os.Args[i]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file "+os.Args[1]+".fog\n %s\n", err.Error()))
	}

	return blkchnFile
}

func readLogFile() string {
	if b, err := ioutil.ReadFile(os.Args[1]); err == nil {
		return string(b)
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse log file.\n %s\n", err.Error()))
		return ""
	}
}

func getBlockLines() (val []string) {
	data := readLogFile()
	s := strings.Split(data, "\n")
	max := 1154482
	if max > len(s) {
		max = len(s)
	}
	if len(s)>1152000 {
		val = make([]string, 0)
		for i:=1152000; i<max; i++ {
			if len(s[i]) > 0 && s[i][0] != '2' {
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

func resolveLevel(i int, blockLevel map[string]int, lines []string, entry []string) int {

	if h, b := blockLevel[entry[2]]; b {
		return h+1
	}

	for j := i+1; j<len(lines); j++ {
		if parentEntry := strings.Split(lines[j], " "); len(parentEntry)>2 && parentEntry[0] != "2" && parentEntry[1] == entry[2] {
			return  resolveLevel(j, blockLevel, lines, parentEntry)+1
		}
	}

	fmt.Println("Error while calculating level for block " + entry[1])

	return 0
}

func parseLog() [][]block {

	blockLevel := make(map[string]int)

	lines := getBlockLines()

	entry := strings.Split(lines[0], " ")

	blockLevel[entry[1]] = -1

	blockChain := make([][]block, len(lines))

	for i:=1; i<len(lines); i++{

		entry = strings.Split(lines[i], " ")

		if entry[0] != "2" && len(entry)>2 {

			h := resolveLevel(i, blockLevel, lines, entry)
			blockLevel[entry[1]] = h

			if blockChain[h] == nil {
				blockChain[h] = make([]block, 0)
			}

			if entry[0]=="0" {
				blockChain[h] = append(blockChain[h], block{Hash:entry[1], Parent:entry[2], Time:time.Unix(0, getNsec(entry[4]))})

			} else {
				blockChain[h] = append(blockChain[h], block{Hash:entry[1], Parent:entry[2], Time:time.Unix(0, getNsec(entry[3]))})
			}
		}
	}

	return blockChain
}

func writeToFile(file *os.File, content string) {
	if _, err := file.Write([]byte(content)); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to %s.\n %s\n", file.Name(), err.Error()))
	}
}

func writeChain(chain [][]block) {

	blkchnFile := createFile(2)

	for i := 0; i<len(chain) && len(chain[i])>0; i++ {
		writeToFile(blkchnFile, chain[i][0].Hash)
		for j:=1; j<len(chain[i]); j++ {
			writeToFile(blkchnFile, " "+chain[i][j].Hash)
		}
		writeToFile(blkchnFile, "\n")
	}
}

func writeHeightTime(chain [][]block) {

	heightFile := createFile(3)

	initTime := chain[0][0].Time

	lastTime := initTime

	var meanDiff int64 = 0

	var i int64

	for i = 0; i<int64(len(chain)) && len(chain[i])>0 && i<1200; i++ {

		s := fmt.Sprintf("%d %s", len(chain[i]), chain[i][0].Time.Format(timeFormat))

		diff := chain[i][0].Time.Sub(initTime)

		s += fmt.Sprintf(" %d:%d:%d ", int64(diff.Hours()), int64(diff.Minutes())%60, int64(diff.Seconds()+0.5)%60)

		diff = chain[i][0].Time.Sub(lastTime)

		meanDiff += diff.Nanoseconds()

		s += fmt.Sprintf("+%d seconds ", int64(diff.Seconds()+0.5))



		if (i>0) {
			s += fmt.Sprintf("- Mean Diff: %d seconds\n", ((meanDiff/(i))+500000000)/1000000000)
		} else {
			s += "\n"
		}

		writeToFile(heightFile, s)


		lastTime = chain[i][0].Time
	}

	//writeToFile(heightFile, fmt.Sprintf("Mean Diff: %d seconds\n", ((meanDiff/i)+500000000)/1000000000))
}


func main(){
	chain := parseLog()

	writeChain(chain)

	writeHeightTime(chain)
}
