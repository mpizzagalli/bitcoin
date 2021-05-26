package nodeEvents

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Nodes events
type NodeEvent struct {
	Event interface{}
}

type BroadcastedBlockEvent struct {
	BlockHash   string
	ParentBlock string
}

type MinedBlockEvent struct {
	BlockHash   string
	ParentBlock string
}

type SelfishMinedBlockEvent struct {
	BlockHash   string
	ParentBlock string
}

// FIXME: Has no error handling
// Reads a file as a string
func readFileLines(filePath string) []string {
	if b, err := ioutil.ReadFile(filePath); err == nil {
		return strings.Split(string(b), "\n")
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to read file %s.\n %s\n", filePath, err.Error()))
		os.Exit(1)
		return nil
	}
}

// Given the lines of logs, filters only the mined block events while parsing them
func readNodeEvents(logFileLines []string) []NodeEvent {
	var nodeEvents []NodeEvent

	for _, logFileLine := range logFileLines {
		var nodeEvent []string = strings.Split(logFileLine, " ")

		if nodeEvent[0] == "0" {
			// Mined block event: 0 | block hash | parent hash | # txs | timestamp
			var minedBlockEvent MinedBlockEvent
			minedBlockEvent.BlockHash = nodeEvent[1]
			minedBlockEvent.ParentBlock = nodeEvent[2]
			nodeEvents = append(nodeEvents, NodeEvent{minedBlockEvent})
		} else if nodeEvent[0] == "1" {
			// Broadcasted block event: 1 | block hash | parent hash | timestamp
			var broadcastedBlockEvent BroadcastedBlockEvent
			broadcastedBlockEvent.BlockHash = nodeEvent[1]
			broadcastedBlockEvent.ParentBlock = nodeEvent[2]
			nodeEvents = append(nodeEvents, NodeEvent{broadcastedBlockEvent})
		} else if nodeEvent[0] == "3" {
			// Selfish mined block event: 3 | block hash | parent hash | # txs | timestamp
			var selfishMinedBlockEvent SelfishMinedBlockEvent
			selfishMinedBlockEvent.BlockHash = nodeEvent[1]
			selfishMinedBlockEvent.ParentBlock = nodeEvent[2]
			nodeEvents = append(nodeEvents, NodeEvent{selfishMinedBlockEvent})
		}
	}

	return nodeEvents
}

func readNodeEventsPizza(logFileLines []string) []NodeEvent {
	var nodeEvents []NodeEvent

	for _, logFileLine := range logFileLines {
		var nodeEvent []string = strings.Split(logFileLine, " ")
		if len(nodeEvent) == 1 {
			continue
		}
		if nodeEvent[1] == "0" {
			// Mined block event: timestamp | 0 | block hash | parent hash | # txs |
			var minedBlockEvent MinedBlockEvent
			minedBlockEvent.BlockHash = nodeEvent[2]
			minedBlockEvent.ParentBlock = nodeEvent[3]
			nodeEvents = append(nodeEvents, NodeEvent{minedBlockEvent})
		} else if nodeEvent[1] == "1" {
			// Broadcasted block event: timestamp |  1 | block hash | parent hash
			var broadcastedBlockEvent BroadcastedBlockEvent
			broadcastedBlockEvent.BlockHash = nodeEvent[2]
			broadcastedBlockEvent.ParentBlock = nodeEvent[3]
			nodeEvents = append(nodeEvents, NodeEvent{broadcastedBlockEvent})
		} else if nodeEvent[1] == "3" {
			// Selfish mined block event: timestamp | 3 | block hash | parent hash | # txs
			var selfishMinedBlockEvent SelfishMinedBlockEvent
			selfishMinedBlockEvent.BlockHash = nodeEvent[2]
			selfishMinedBlockEvent.ParentBlock = nodeEvent[3]
			nodeEvents = append(nodeEvents, NodeEvent{selfishMinedBlockEvent})
		}
	}

	return nodeEvents
}

func LoadNodeEvents(pathToNodeLogs string) []NodeEvent {
	var logs []string = readFileLines(pathToNodeLogs)

	var nodeEvents = readNodeEvents(logs)
	return nodeEvents
}

func LoadNodeEventsPizza(pathToNodeLogs string) []NodeEvent {
	var logs []string = readFileLines(pathToNodeLogs)

	var nodeEvents = readNodeEventsPizza(logs)
	return nodeEvents
}
