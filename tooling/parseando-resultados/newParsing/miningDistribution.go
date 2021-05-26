package newParsing

import (
	"fmt"
	"strconv"

	NodeEvents "../nodeEvents"
)

type BlockInfo struct {
	BlockHash   string
	ParentBlock string
	// -1 represents that the miner is unknown
	MinerNode int
	//Please note that this might be just in regard to the processed file and not the whole chain
	Height      int
	InMainChain bool
	Public      bool
}

type SingleNodeMiningDistribution struct {
	Own   int
	Total int
}

type SingleNodeMetric struct {
	Mined        int
	Accepted     int
	Ignored      int
	NotPublished int
}

func getPathToNodeLogs(folderWithNodesLogs string, nodeNumber int) string {
	var nodeNumberAsString = strconv.Itoa(nodeNumber)
	var pathToNodeLogs = folderWithNodesLogs + "/btcCoreLogN" + nodeNumberAsString

	return pathToNodeLogs
}

func getNodeBlockTreeFromEvents(nodeEvents []NodeEvents.NodeEvent) map[string]BlockInfo {
	const minedByOwn = 0
	const minedByOthers = 1500
	blocksByHash := make(map[string]BlockInfo)

	// Get the initial map of blocks by hash (without height)
	for _, nodeEvent := range nodeEvents {
		// fmt.Println("new node Event received:")
		switch event := nodeEvent.Event.(type) {
		case NodeEvents.MinedBlockEvent:
			// fmt.Println("My Block")
			var blockInfo BlockInfo
			blockInfo.BlockHash = event.BlockHash
			blockInfo.ParentBlock = event.ParentBlock
			blockInfo.MinerNode = minedByOwn
			blockInfo.Height = -1
			blockInfo.InMainChain = false
			blockInfo.Public = true

			blocksByHash[event.BlockHash] = blockInfo
		case NodeEvents.SelfishMinedBlockEvent: // This means it was published publicly?
			// fmt.Println("My selfish Block")
			var blockInfo BlockInfo
			blockInfo.BlockHash = event.BlockHash
			blockInfo.ParentBlock = event.ParentBlock
			blockInfo.MinerNode = minedByOwn
			blockInfo.Height = -1
			blockInfo.InMainChain = false
			blockInfo.Public = false

			blocksByHash[event.BlockHash] = blockInfo
		case NodeEvents.BroadcastedBlockEvent:
			// fmt.Println("Someone else Block")
			var blockInfo BlockInfo
			blockInfo.BlockHash = event.BlockHash
			blockInfo.ParentBlock = event.ParentBlock
			blockInfo.MinerNode = minedByOthers
			blockInfo.Height = -1
			blockInfo.InMainChain = false
			blockInfo.Public = true

			blocksByHash[event.BlockHash] = blockInfo
		}
		// fmt.Println(nodeEvent)
	}

	// Calculate the height of each block
	for blockHash, _ := range blocksByHash {
		updateHeightOfBlockAndAncestors(blockHash, blocksByHash)
	}
	return blocksByHash
}

func updateHeightOfBlockAndAncestors(blockHash string, allBlocks map[string]BlockInfo) {
	var blockInfo = allBlocks[blockHash]

	var parentBlockHash = blockInfo.ParentBlock
	parentBlockInfo, containsParentBlock := allBlocks[parentBlockHash]

	if blockInfo.Height >= 0 {
	} else if !containsParentBlock {
		blockInfo.Height = 0
	} else if parentBlockInfo.Height >= 0 {
		blockInfo.Height = parentBlockInfo.Height + 1
	} else {
		updateHeightOfBlockAndAncestors(parentBlockHash, allBlocks)
		blockInfo.Height = allBlocks[parentBlockHash].Height + 1
	}
	allBlocks[blockHash] = blockInfo
}

func getAndMarkAncestorsOfMainChain(blockHash string, allBlocks map[string]BlockInfo) []BlockInfo {
	var blockInfo = allBlocks[blockHash]
	blockInfo.InMainChain = true
	allBlocks[blockHash] = blockInfo
	var parentBlockHash = blockInfo.ParentBlock
	if _, isParentBlockContained := allBlocks[parentBlockHash]; isParentBlockContained {
		var ancestorsOfParent = getAndMarkAncestorsOfMainChain(parentBlockHash, allBlocks)
		return append(ancestorsOfParent, blockInfo)
	} else {
		return append([]BlockInfo{}, blockInfo)
	}
}

func getAndMarkMainChain(blocksByHash map[string]BlockInfo) []BlockInfo {
	// Obtain the block with the biggest height
	var highestHeight = -1
	var blockWithHighestHeight = ""
	for blockHash, blockInfo := range blocksByHash {
		if blockInfo.Height > highestHeight {
			highestHeight = blockInfo.Height
			blockWithHighestHeight = blockHash
		}
	}

	if blockWithHighestHeight == "" {
		return nil
	} else {
		return getAndMarkAncestorsOfMainChain(blockWithHighestHeight, blocksByHash)
	}
}

func getMinerMetrics(blocksByHash map[string]BlockInfo) map[int]SingleNodeMetric {
	// minerDistribution := make(map[int]int)
	metrics := make(map[int]SingleNodeMetric)

	for _, blockInfo := range blocksByHash {
		// minerDistribution[blockInfo.MinerNode] = minerDistribution[blockInfo.MinerNode] + 1
		fmt.Println(blockInfo)
		miner := blockInfo.MinerNode
		minerMetric, foo := metrics[miner]
		if !foo {
			fmt.Println("First time with miner: ", miner)
			var newMinerMetric SingleNodeMetric
			minerMetric = newMinerMetric
		}
		minerMetric.Mined += 1
		if !blockInfo.Public {
			minerMetric.NotPublished += 1
		} else if blockInfo.InMainChain {
			minerMetric.Accepted += 1
		} else {
			minerMetric.Ignored += 1
		}
		fmt.Println(minerMetric)
		metrics[miner] = minerMetric
	}
	return metrics
}

func CalculateMetricFromSelfishPOV(folderWithNodesLogs string, nodeNumber int) SingleNodeMiningDistribution {
	var pathToNodeLogs = getPathToNodeLogs(folderWithNodesLogs, nodeNumber)
	var nodeEvents = NodeEvents.LoadNodeEventsPizza(pathToNodeLogs)

	var nodeBlockTree = getNodeBlockTreeFromEvents(nodeEvents)
	fmt.Println(nodeBlockTree)
	var nodeMainChain = getAndMarkMainChain(nodeBlockTree)
	fmt.Println(nodeMainChain)
	var minerDistribution = getMinerMetrics(nodeBlockTree)

	fmt.Println(minerDistribution)

	return SingleNodeMiningDistribution{0, 0}
}
