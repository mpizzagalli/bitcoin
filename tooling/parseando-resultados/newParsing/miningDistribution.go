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
	Mined            int
	Accepted         int
	Ignored          int
	NotPublished     int
	HashRate         float64
	MinerRate        float64
	AcceptanceRate   float64
	IgnoranceRate    float64
	NotPublishedRate float64
}

func getPathToNodeLogs(folderWithNodesLogs string, nodeNumber int) string {
	var nodeNumberAsString = strconv.Itoa(nodeNumber)
	var pathToNodeLogs = folderWithNodesLogs + "/btcCoreLogN" + nodeNumberAsString

	return pathToNodeLogs
}

func getNodeBlockTreeForSinglePOV(nodeEvents []NodeEvents.NodeEvent) map[string]BlockInfo {
	const minedByOwn = 0
	const minedByOthers = -1
	blocksByHash := make(map[string]BlockInfo)

	// Get the initial map of blocks by hash (without height)
	for _, nodeEvent := range nodeEvents {
		var blockInfo BlockInfo
		switch event := nodeEvent.Event.(type) {
		case NodeEvents.MinedBlockEvent:
			blockInfo.BlockHash = event.BlockHash
			blockInfo.ParentBlock = event.ParentBlock
			blockInfo.MinerNode = minedByOwn
			blockInfo.Height = -1
			blockInfo.InMainChain = false
			blockInfo.Public = true

			blocksByHash[event.BlockHash] = blockInfo
		case NodeEvents.SelfishMinedBlockEvent: // This means it was published publicly?
			blockInfo.BlockHash = event.BlockHash
			blockInfo.ParentBlock = event.ParentBlock
			blockInfo.MinerNode = minedByOwn
			blockInfo.Height = -1
			blockInfo.InMainChain = false
			blockInfo.Public = false

			blocksByHash[event.BlockHash] = blockInfo
		case NodeEvents.BroadcastedBlockEvent:
			blockInfo.BlockHash = event.BlockHash
			blockInfo.ParentBlock = event.ParentBlock
			blockInfo.MinerNode = minedByOthers
			blockInfo.Height = -1
			blockInfo.InMainChain = false
			blockInfo.Public = true

			blocksByHash[event.BlockHash] = blockInfo
		case NodeEvents.BroadcastedHeadersBlockEvent:
			blockInfo.BlockHash = event.BlockHash
			blockInfo.MinerNode = minedByOthers
			blockInfo.Height = -1
			blockInfo.InMainChain = false
			blockInfo.Public = true
			if _, exists := blocksByHash[event.BlockHash]; !exists {
				blocksByHash[event.BlockHash] = blockInfo
			}
		}
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
	fmt.Println(highestHeight)
	if blockWithHighestHeight == "" {
		return nil
	} else {
		return getAndMarkAncestorsOfMainChain(blockWithHighestHeight, blocksByHash)
	}
}

func getMinerMetrics(blocksByHash map[string]BlockInfo) map[string]SingleNodeMetric {
	metrics := make(map[string]SingleNodeMetric)
	var totalMetrics SingleNodeMetric
	for _, blockInfo := range blocksByHash {
		miner := blockInfo.MinerNode
		minerMetric := metrics[strconv.Itoa(miner)]
		minerMetric.Mined += 1
		totalMetrics.Mined += 1
		if !blockInfo.Public {
			minerMetric.NotPublished += 1
			totalMetrics.NotPublished += 1
		} else if blockInfo.InMainChain {
			minerMetric.Accepted += 1
			totalMetrics.Accepted += 1
		} else {
			minerMetric.Ignored += 1
			totalMetrics.Ignored += 1
		}
		metrics[strconv.Itoa(miner)] = minerMetric
	}
	metrics["total"] = totalMetrics
	for id, metric := range metrics {
		metric.HashRate = float64(metric.Mined) / float64(totalMetrics.Mined)
		metric.MinerRate = float64(metric.Accepted) / float64(totalMetrics.Accepted)
		metric.AcceptanceRate = float64(metric.Accepted) / float64(metric.Mined)
		metric.IgnoranceRate = float64(metric.Ignored) / float64(metric.Mined)
		metric.NotPublishedRate = float64(metric.NotPublished) / float64(metric.Mined)
		metrics[id] = metric
	}
	return metrics
}

func CalculateMetricForSinglePOV(folderWithNodesLogs string, nodeNumber int) map[string]SingleNodeMetric {
	var pathToNodeLogs = getPathToNodeLogs(folderWithNodesLogs, nodeNumber)
	var nodeEvents = NodeEvents.LoadNodeEventsPizza(pathToNodeLogs)

	var nodeBlockTree = getNodeBlockTreeForSinglePOV(nodeEvents)
	getAndMarkMainChain(nodeBlockTree)
	var minerMetrics = getMinerMetrics(nodeBlockTree)

	fmt.Println(minerMetrics)

	return minerMetrics
}

func CalculateMetrics(folderWithNodesLogs string, nodeAmount int) map[string]SingleNodeMetric {
	var nodeEvents = loadNodesEvents(nodeAmount, folderWithNodesLogs)

	var nodeBlockTree = getNodeBlockTree(nodeEvents)
	getAndMarkMainChain(nodeBlockTree)
	var minerMetrics = getMinerMetrics(nodeBlockTree)

	fmt.Println(minerMetrics)

	return minerMetrics
}

func loadNodesEvents(nodeAmount int, folderWithNodesLogs string) [][]NodeEvents.NodeEvent {
	var eventsPerNode = [][]NodeEvents.NodeEvent{}
	for i := 0; i < nodeAmount; i++ {
		var pathToNodeLogs = getPathToNodeLogs(folderWithNodesLogs, i)
		var nodeEvents = NodeEvents.LoadNodeEventsPizza(pathToNodeLogs)
		eventsPerNode = append(eventsPerNode, nodeEvents)
	}
	return eventsPerNode
}

func getNodeBlockTree(eventsPerNode [][]NodeEvents.NodeEvent) map[string]BlockInfo {
	blocksByHash := make(map[string]BlockInfo)

	// Get the initial map of blocks by hash (without height)
	for nodeNumber, logsBlockInfo := range eventsPerNode {
		for _, nodeEvent := range logsBlockInfo {
			var blockInfo BlockInfo
			switch event := nodeEvent.Event.(type) {
			case NodeEvents.MinedBlockEvent:
				blockInfo.BlockHash = event.BlockHash
				blockInfo.ParentBlock = event.ParentBlock
				blockInfo.MinerNode = nodeNumber
				blockInfo.Height = -1
				blockInfo.InMainChain = false
				blockInfo.Public = true

				blocksByHash[event.BlockHash] = blockInfo
			case NodeEvents.SelfishMinedBlockEvent:
				blockInfo.BlockHash = event.BlockHash
				blockInfo.ParentBlock = event.ParentBlock
				blockInfo.MinerNode = nodeNumber
				blockInfo.Height = -1
				blockInfo.InMainChain = false
				blockInfo.Public = false

				blocksByHash[event.BlockHash] = blockInfo
			}
		}
	}

	// Calculate the height of each block
	for blockHash, _ := range blocksByHash {
		updateHeightOfBlockAndAncestors(blockHash, blocksByHash)
	}
	return blocksByHash
}
