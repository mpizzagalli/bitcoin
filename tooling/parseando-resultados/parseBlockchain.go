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

// Creates a file using the i-th program parameter as it's path
func createFile(i int) (blkchnFile *os.File) {
	var err error

	if blkchnFile, err = os.Create(os.Args[i]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file "+os.Args[1]+".fog\n %s\n", err.Error()))
	}

	return blkchnFile
}

// Returns the string of the passed log file (first parameter)
func readLogFile() string {
	if b, err := ioutil.ReadFile(os.Args[1]); err == nil {
		return string(b)
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse log file.\n %s\n", err.Error()))
		return ""
	}
}

// Returns between 1152000 (if not it doesn't do anything) and 1154482 (remainder are ignored)
// lines from the log file
func getBlockLines() (val []string) {
	data := readLogFile()
	s := strings.Split(data, "\n")
	max := 1154482
	if max > len(s) {
		max = len(s)
	}
	// Note that if this min is changed then it should be at least of value 1 due to "Starting bitcoin client at"
	const min = 1152000
	if len(s)>min {
		val = make([]string, 0)
		for i:=min; i<max; i++ {
			if len(s[i]) > 0 && s[i][0] != '2' {
				buff := make([]byte, len(s[i]))
				_ = copy(buff, s[i])
				val = append(val, string(buff))
			}
		}
	}
	return
}

// Parses a string number to a int64
func getNsec(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// Returns the block level of a given entry
// FIXME: This is really inefficient as it doesn't save useful block level calculations
func resolveLevel(i int, blockLevel map[string]int, lines []string, entry []string) int {

	// If the block level of the parent was already calculated just return 1 plus it
	if h, b := blockLevel[entry[2]]; b {
		return h+1
	}

	// Searches for the parent of the entry passed
	// FIXME: The parent entry should be backwards... why does this search forward?
	//		  If properly called this line is never reached as the parent should have been already parsed
	for j := i+1; j<len(lines); j++ {
		if parentEntry := strings.Split(lines[j], " "); len(parentEntry)>2 && parentEntry[0] != "2" && parentEntry[1] == entry[2] {
			return resolveLevel(j, blockLevel, lines, parentEntry)+1
		}
	}

	fmt.Println("Error while calculating level for block " + entry[1])

	return 0
}

// Obtains some of the entries in the parameter log file sorted by "block level"
func parseLog() [][]block {

	// Map: block hash -> block level (proportional to block number)
	blockLevel := make(map[string]int)

	lines := getBlockLines()

	// FIXME: If there're less lines than the minimum this fails as lines is empty
	entry := strings.Split(lines[0], " ")

	blockLevel[entry[1]] = -1

	blockChain := make([][]block, len(lines))

	for i:=1; i<len(lines); i++{

		entry = strings.Split(lines[i], " ")

		// Parse only if it's a log of a mined/received block
		if entry[0] != "2" && len(entry)>2 {

			h := resolveLevel(i, blockLevel, lines, entry)
			blockLevel[entry[1]] = h

			if blockChain[h] == nil {
				blockChain[h] = make([]block, 0)
			}

			// Creates an entry with the line data for returning
			if entry[0]=="0" {
				blockChain[h] = append(blockChain[h], block{Hash:entry[1], Parent:entry[2], Time:time.Unix(0, getNsec(entry[4]))})
			} else {
				blockChain[h] = append(blockChain[h], block{Hash:entry[1], Parent:entry[2], Time:time.Unix(0, getNsec(entry[3]))})
			}
		}
	}

	return blockChain
}

// Writes a line to the passed file
func writeToFile(file *os.File, content string) {
	if _, err := file.Write([]byte(content)); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to %s.\n %s\n", file.Name(), err.Error()))
	}
}

// Writes the hashes of the blocks on a file separated by spaces, one line per level
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

// Parses the first block of at most 1200 levels of the blockchain, writing into a file:
//		timestamp | time since first level | time since previous level | mean diff between levels (till this one)
func writeHeightTime(chain [][]block) {

	heightFile := createFile(3)

	initTime := chain[0][0].Time

	lastTime := initTime

	var meanDiff int64 = 0

	var i int64

	// This parses at most the first 1200 levels of the blockchain, with only the first entry from each
	for i = 0; i<int64(len(chain)) && len(chain[i])>0 && i<1200; i++ {

		// s << timestamp of the first block of the level
		s := fmt.Sprintf("%d %s", len(chain[i]), chain[i][0].Time.Format(timeFormat))

		// diff <- time since the first entry
		diff := chain[i][0].Time.Sub(initTime)

		// s << time since the first entry
		s += fmt.Sprintf(" %d:%d:%d ", int64(diff.Hours()), int64(diff.Minutes())%60, int64(diff.Seconds()+0.5)%60)

		// diff <- time since the previous entry
		diff = chain[i][0].Time.Sub(lastTime)

		// meanDiff <- accumlator of diferences
		meanDiff += diff.Nanoseconds()

		// s << time since the previous entry
		s += fmt.Sprintf("+%d seconds ", int64(diff.Seconds()+0.5))



		if (i>0) {
			// s << mean difference between blocks
			s += fmt.Sprintf("- Mean Diff: %d seconds\n", ((meanDiff/(i))+500000000)/1000000000)
		} else {
			s += "\n"
		}

		writeToFile(heightFile, s)


		lastTime = chain[i][0].Time
	}

	//writeToFile(heightFile, fmt.Sprintf("Mean Diff: %d seconds\n", ((meanDiff/i)+500000000)/1000000000))
}


/*
	Parameters:
	- Input node log file (btcCoreLogN_)
	- Output block hashes file (by level)
	- Output processed block timestamps file (by level)
 */
func main(){
	chain := parseLog()

	writeChain(chain)

	writeHeightTime(chain)
}
