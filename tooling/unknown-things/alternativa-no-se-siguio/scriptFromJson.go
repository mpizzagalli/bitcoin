package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"time"
	"sort"
	"fmt"
	"os/exec"
)

type Country struct {
	Id string `json:"id"`
	Nodes []int `json:"nodes"`
	LocalLatency int `json:"intra_latency"`
}

type Latency struct {
	A string `json:"a"`
	B string `json:"b"`
	Latency int `json:"latency"`
}

type countriesJson struct {
	Countries []Country `json:"countries"`
    Latencies []Latency `json:"inter_latencies"`
}

type BitcoinNode struct {
	Id int `json:"id"`
	Peers []int `json:"peers"`
	GeneratesBlocks bool `json:"is_miner"`
	CreatesTransactions bool `json:"creates_txs"`
}

type JsonEvent struct {
	MsOffset uint32 `json:"time"`
	Node uint32 `json:"node_id"`
}

type EventsFile struct {
	Blocks []JsonEvent `json:"blocks"`
	Transactions []JsonEvent `json:"transactions"`
}

const baseEventFilename string = "logN"

type KindOfEvent byte

const (
	Block KindOfEvent = 0
	Tx KindOfEvent = 1
)

var hostIps = [4][]byte{[]byte("0.1.10.162\n"), []byte("0.1.10.163\n"), []byte("0.1.10.166\n"), []byte("0.1.10.167\n")}
var numberOfHosts int

type NodeEvent struct {
	Offset time.Duration
	Type KindOfEvent
}

type NodeLog []NodeEvent

func (s NodeLog) Len() int {
	return len(s)
}
func (s NodeLog) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s NodeLog) Less(i, j int) bool {
	return s[i].Offset < s[j].Offset
}

func writeLineToFile(file *os.File, content string) {
	if _, err := file.Write([]byte(content+"\n")); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to file.\n %s\n", err.Error()))
		return
	}
}

func createFile() (scriptFile *os.File) {
	var err error

	if scriptFile, err = os.Create(os.Args[1]+".fog"); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file.\n %s\n", err.Error()))
	}

	return scriptFile
}

func makePhysicalLayer(scriptFile *os.File, nodeIdToHost map[int]int) {

	var data countriesJson

	if jsonBytes, err := ioutil.ReadFile(os.Args[2]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse physical layer file.\n %s\n", err.Error()))
		return
	} else if err = json.Unmarshal(jsonBytes, &data); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse physical layer json.\n %s\n", err.Error()))
		return
	}

	writeLineToFile(scriptFile ,fmt.Sprintf("for i in 0..%d do", len(data.Countries)))
	writeLineToFile(scriptFile ,"\tdef n{i}\nend for\n")

	countryIdToHost := make(map[string]int)
	var nextHost int = len(data.Countries)
	var nodeNumber string
	var node int

	for i := 0; i<len(data.Countries); i++ {
		for _, node = range data.Countries[i].Nodes {
			nodeNumber = strconv.Itoa(nextHost)
			writeLineToFile(scriptFile,fmt.Sprintf("def n%s", nodeNumber))
			writeLineToFile(scriptFile,fmt.Sprintf("run n%s netns n%s bash /home/mgeier/ndecarli/invokeBitcoin.sh %s",nodeNumber,nodeNumber,nodeNumber))
			writeLineToFile(scriptFile,fmt.Sprintf("connect n%d n%s %dms\n",i,nodeNumber,data.Countries[i].LocalLatency))
			nodeIdToHost[node] = nextHost
			nextHost++
		}
		countryIdToHost[data.Countries[i].Id] = i
	}

	numberOfHosts = nextHost
	
	for _, lat := range data.Latencies {
		writeLineToFile(scriptFile ,fmt.Sprintf("connect n%d n%d %dms", countryIdToHost[lat.A], countryIdToHost[lat.B], lat.Latency))
	}

	writeLineToFile(scriptFile, "\nbuild-network\n")
}

func makeLogicalLayer(scriptFile *os.File, nodeIdToHost map[int]int) {

	var nodes []BitcoinNode

	if jsonBytes, err := ioutil.ReadFile(os.Args[3]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse logical layer file.\n %s\n", err.Error()))
		return
	} else if err = json.Unmarshal(jsonBytes, &nodes); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse logical layer json.\n %s\n", err.Error()))
		return
	}

	var hostA string
	var nodeA string
	var nodeB int

	for _, from := range nodes {
		hostA = strconv.Itoa(nodeIdToHost[from.Id])
		nodeA = strconv.Itoa(from.Id)
		for _, nodeB = range from.Peers {
			writeLineToFile(scriptFile,fmt.Sprintf("run n%s netns n%s bash /home/mgeier/ndecarli/connectNodes.sh %s %d n%d", hostA, hostA, nodeA, nodeB, nodeIdToHost[nodeB]))
		}
	}

	writeLineToFile(scriptFile, "") //endl para mayor legibilidad
}

func eventFilename(node uint32) string {
	return baseEventFilename + strconv.FormatUint(uint64(node), 10)
}

func makeEvents(scriptFile *os.File, nodeIdToHost map[int]int) {

	var events EventsFile

	if jsonBytes, err := ioutil.ReadFile(os.Args[4]); err != nil {
		os.Stderr.WriteString("Failed to parse events file.\n " + err.Error()+"\n")
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse events file.\n %s\n", err.Error()))
		return
	} else if err = json.Unmarshal(jsonBytes, &events); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse events json.\n %s\n", err.Error()))
		return
	}

	eventsPool := make(map[uint32]NodeLog)

	var newEvent NodeEvent

	newEvent.Type = Block

	for i:=0; i < len(events.Blocks); i++ {

		newEvent.Offset = time.Duration(events.Blocks[i].MsOffset) * time.Millisecond

		if eventsPool[events.Blocks[i].Node] == nil {
			eventsPool[events.Blocks[i].Node] = []NodeEvent{newEvent}
		} else {
			eventsPool[events.Blocks[i].Node] = append(eventsPool[events.Blocks[i].Node], newEvent)
		}
	}

	newEvent.Type = Tx

	for i:=0; i < len(events.Transactions); i++ {

		newEvent.Offset = time.Duration(events.Transactions[i].MsOffset) * time.Millisecond

		if eventsPool[events.Transactions[i].Node] == nil {
			eventsPool[events.Transactions[i].Node] = []NodeEvent{newEvent}
		} else {
			eventsPool[events.Transactions[i].Node] = append(eventsPool[events.Transactions[i].Node], newEvent)
		}
	}

	var eventsFile *os.File
	var fileErr error

	for k,v := range eventsPool {
		sort.Sort(v)
		filename := eventFilename(k)

		if eventsFile, fileErr = os.Create(filename); fileErr != nil {
			os.Stderr.WriteString(fmt.Sprintf("Failed to create node events file.\n %s\n", fileErr.Error()))
			return
		} else {
			for i:=0; i<len(v); i++ {
				writeLineToFile(eventsFile, fmt.Sprintf("%d %s", v[i].Type, v[i].Offset.String()))
			}
		}

		shost := strconv.Itoa(nodeIdToHost[int(k)])
		writeLineToFile(scriptFile, fmt.Sprintf("run n%s netns n%s scp -o StrictHostKeyChecking=no root@admgnode2:/home/mgeier/ndecarli/%s /home/mgeier/ndecarli/%s", shost, shost, filename, filename))
	}

	writeLineToFile(scriptFile, "")

	startTime := time.Now().Add(time.Minute * 10)
	//TODO: get timestamp in script using date +'%s %N'

	for k, _:= range eventsPool {
		shost := strconv.Itoa(nodeIdToHost[int(k)])
		writeLineToFile(scriptFile,fmt.Sprintf("run n%s netns n%s go run /home/mgeier/ndecarli/schedTest.go %s %d %d & disown", shost, shost, eventFilename(k), k, startTime.Unix()))
	}
	
	testDuration := events.Transactions[len(events.Transactions)-1].MsOffset

	if latestBlockEvent := events.Blocks[len(events.Blocks)-1].MsOffset; testDuration < latestBlockEvent {
		testDuration = latestBlockEvent
	}

	testDuration += 600000 // agregamos 10 min para que todo se termine de propagar y asentar en disco

	writeLineToFile(scriptFile, fmt.Sprintf("\nrun n0 sleep %dms\n", testDuration))

	for node, host := range nodeIdToHost {
		shost := strconv.Itoa(host)
		writeLineToFile(scriptFile, fmt.Sprintf("run n%s netns n%s bash /home/mgeier/ndecarli/bitcoindo.sh %d stop", shost, shost, node))
	}

	if fileErr = scriptFile.Close(); fileErr == nil {
		ipsFilename := os.Args[1]+"ips.txt"
		if ipsFile, err := os.Create(ipsFilename); err == nil {
			for i:=0; i<numberOfHosts && err == nil; i++ {
				_, err = ipsFile.Write(hostIps[i&3])
			}
			if err == nil {
				launchFog := exec.Command("python3", "/home/mgeier/repos/sherlockfog", "/home/mgeier/ndecarli/"+scriptFile.Name(), "/home/mgeier/ndecarli/"+ipsFilename, "&", "disown")
				if err = launchFog.Run(); err != nil {
					os.Stderr.WriteString(fmt.Sprintf("Failed to launch sherlock fog.\n %s\n", err.Error()))
				}
			} else {
				os.Stderr.WriteString(fmt.Sprintf("Failed to write ips file.\n %s\n", err.Error()))
			}
		} else {
			os.Stderr.WriteString(fmt.Sprintf("Failed to create ips file.\n %s\n", err.Error()))
		}
	} else {
		os.Stderr.WriteString("Could not close fog file properly.\n")
	}
	
}



func main(){

	if len(os.Args) < 5 {
		os.Stderr.WriteString("Missing Json filenames as arguments.\n")
		return
	}

	scriptFile := createFile()

	nodeIdToHost := make(map[int]int)

	makePhysicalLayer(scriptFile, nodeIdToHost)

	makeLogicalLayer(scriptFile, nodeIdToHost)

	makeEvents(scriptFile, nodeIdToHost)
}

