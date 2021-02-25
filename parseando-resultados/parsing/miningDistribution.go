package parsing

import (
	"os"
	"fmt"
	"math"
	"strconv"
	NodeEvents "../nodeEvents"
)

type BlockInfo struct {
	BlockHash string
	ParentBlock string
	// -1 represents that the miner is unknown
	MinerNode int
	//Please note that this might be just in regard to the processed file and not the whole chain
	Height int
}


/** Practical from all nodes **/
func CalculateTentativeMiningDistribution(numberOfNodes int, folderWithNodeLogs string) map[int]int {
	var eventsPerNode = loadNodesEvents(numberOfNodes, folderWithNodeLogs)
	var tentativeBlockTree = getTentativeBlockTreeFromEvents(eventsPerNode)
	var tentativeMainChain = getMainChain(tentativeBlockTree)
	var tentativeMinerDistribution = getMinerDistribution(tentativeMainChain)
	return tentativeMinerDistribution
}

func CalculateAndLogMiningDistribution(numberOfNodes int, folderWithNodesLogs string) {
	var eventsPerNode = loadNodesEvents(numberOfNodes, folderWithNodesLogs)

	// Main chain statistics
	var blockTree = getBlockTreeFromEvents(eventsPerNode)
	var mainChain = getMainChain(blockTree)

	var minerDistribution = getMinerDistribution(mainChain)

	logMinerDistribution(minerDistribution, "Obtained")
	os.Stdout.WriteString("\n")

	// Tentative main chain statistics
	var tentativeBlockTree = getTentativeBlockTreeFromEvents(eventsPerNode)
	var tentativeMainChain = getMainChain(tentativeBlockTree)

	var tentativeMinerDistribution = getMinerDistribution(tentativeMainChain)

	logMinerDistribution(tentativeMinerDistribution, "Tentative")
}

func CalculateMiningDistributionInIntervals(numberOfNodes int, folderWithNodesLogs string, step int) {
	var eventsPerNode = loadNodesEvents(numberOfNodes, folderWithNodesLogs)

	var tentativeBlockTree = getTentativeBlockTreeFromEvents(eventsPerNode)
	var mainChain = getMainChain(tentativeBlockTree)

	for lastIndex := step; lastIndex < len(mainChain); lastIndex += step {
		var chain = mainChain[0:lastIndex]

		var minerDistribution = getMinerDistribution(chain)

		logMinerDistribution(minerDistribution, "Tentative prefix")
		os.Stdout.WriteString("\n")
	}

}

// Experiments constants
const selfishMiner = 1
var alphas [6]float64 = [6]float64{0.1, 0.2, 0.3, 0.4, 0.45, 0.475}
var deltas [3]float64 = [3]float64{0, 0.5, 1}
func CalculateRewardDistributionExperiments(folderWithExperiments string) {
	for _, alpha := range alphas {
		for _,  delta := range deltas {
			var experimentResultPath = fmt.Sprintf("%s/sm-%f-%f", folderWithExperiments, alpha, delta)
			os.Stdout.WriteString(fmt.Sprintf("alpha: %f, delta: %f\n", alpha, delta))

			// Log tentatively obtained from experiments
			var minerDistribution = CalculateTentativeMiningDistribution(3, experimentResultPath)
			logMinerDistribution(minerDistribution, "Obtained tentative")

			// Log expected from experiments
			var expectedSelfishMiningProportion = CalculateExpectedSelfishMiningDistribution(alpha, delta)
			os.Stdout.WriteString(fmt.Sprintf("Expected: %f\n",  expectedSelfishMiningProportion))
			os.Stdout.WriteString("\n")
		} 
	}
}

func CalculateSelfishMinerProportionExperiments(folderWithExperiments string) {
	// Log header
	os.Stdout.WriteString("\t")
	for _, alpha := range alphas {
		os.Stdout.WriteString(fmt.Sprintf("%f\t", alpha))
	}
	os.Stdout.WriteString("\n")

	for _,  delta := range deltas {
		os.Stdout.WriteString(fmt.Sprintf("%f - obtained\t", delta))
		for _, alpha := range alphas {
			var experimentResultPath = fmt.Sprintf("%s/sm-%f-%f", folderWithExperiments, alpha, delta)

			// Calculate proportion mined by selfish
			var minerDistribution = CalculateTentativeMiningDistribution(3, experimentResultPath)
			var selfishMinedBlocks = minerDistribution[selfishMiner]
			var totalMinedBlocks = 0
			for _, numberMinedBlocks := range minerDistribution {
				totalMinedBlocks += numberMinedBlocks
			}
			var selfishMinerProportion = float64(selfishMinedBlocks) / float64(totalMinedBlocks)

			os.Stdout.WriteString(fmt.Sprintf("%f\t", selfishMinerProportion))
		}
		os.Stdout.WriteString("\n")
	}

	for _,  delta := range deltas {
		os.Stdout.WriteString(fmt.Sprintf("%f - expected\t", delta))
		for _, alpha := range alphas {
			// Log expected from experiments
			var expectedSelfishMiningProportion = CalculateExpectedSelfishMiningDistribution(alpha, delta)
			os.Stdout.WriteString(fmt.Sprintf("%f\t",  expectedSelfishMiningProportion))
		}
		os.Stdout.WriteString("\n")
	}
}

// Returns the block tree from mined block event
func getBlockTreeFromEvents(eventsPerNode [][]NodeEvents.NodeEvent) map[string]BlockInfo {
	blocksByHash := make(map[string]BlockInfo)

	// Get the initial map of blocks by hash (without height)
	for nodeNumber, logsBlockInfo := range eventsPerNode {
		for _, nodeEvent := range logsBlockInfo {
			switch event := nodeEvent.Event.(type) {
			case NodeEvents.MinedBlockEvent:
				var blockInfo BlockInfo
				blockInfo.BlockHash = event.BlockHash
				blockInfo.ParentBlock = event.ParentBlock
				blockInfo.MinerNode = nodeNumber
				blockInfo.Height = -1
	
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

// Returns the tentative block tree from mined and selfish mined block event
func getTentativeBlockTreeFromEvents(eventsPerNode [][]NodeEvents.NodeEvent) map[string]BlockInfo {
	blocksByHash := make(map[string]BlockInfo)

	// Get the initial map of blocks by hash (without height)
	for nodeNumber, logsBlockInfo := range eventsPerNode {
		for _, nodeEvent := range logsBlockInfo {
			switch event := nodeEvent.Event.(type) {
			case NodeEvents.MinedBlockEvent:
				var blockInfo BlockInfo
				blockInfo.BlockHash = event.BlockHash
				blockInfo.ParentBlock = event.ParentBlock
				blockInfo.MinerNode = nodeNumber
				blockInfo.Height = -1
	
				blocksByHash[event.BlockHash] = blockInfo
			case NodeEvents.SelfishMinedBlockEvent:
				var blockInfo BlockInfo
				blockInfo.BlockHash = event.BlockHash
				blockInfo.ParentBlock = event.ParentBlock
				blockInfo.MinerNode = nodeNumber
				blockInfo.Height = -1
	
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

func logMinerDistribution(minerDistribution map[int]int, prefix string) {
	var totalMinedBlocks = 0
	for _, numberMinedBlocks := range minerDistribution {
		totalMinedBlocks += numberMinedBlocks
	}

	os.Stdout.WriteString(fmt.Sprintf("%s, total mined blocks: %d\n", prefix, totalMinedBlocks))
	for node, numberMinedBlocks := range minerDistribution {
		var percentageMinedByNode = float32(numberMinedBlocks) / float32(totalMinedBlocks)
		os.Stdout.WriteString(fmt.Sprintf("Node %d mined: %d (%f of total)\n", node, numberMinedBlocks, percentageMinedByNode))
	}
}

func loadNodesEvents(numberOfNodes int, folderWithNodesLogs string) [][]NodeEvents.NodeEvent {
	var eventsPerNode = [][]NodeEvents.NodeEvent{}
	for i:=0; i<numberOfNodes; i++ {
		var pathToNodeLogs = getPathToNodeLogs(folderWithNodesLogs, i)
		var nodeEvents = NodeEvents.LoadNodeEvents(pathToNodeLogs)
		eventsPerNode = append(eventsPerNode, nodeEvents)
	}
	return eventsPerNode
}


/** Theorical **/
func CalculateExpectedSelfishMiningDistribution(alpha float64, delta float64) float64 {
	var dividend = alpha * math.Pow(1 - alpha, 2) * (4 * alpha + delta * (1 - 2 * alpha)) - math.Pow(alpha, 3)
	var divisor = 1 - alpha * (1 + (2 - alpha) * alpha)
	return dividend / divisor
}


/** Practical from a single node **/
type SingleNodeMiningDistribution struct {
	Own int
	Total int
}

func CalculateMiningDistributionFromNodePerspective(folderWithNodesLogs string, nodeNumber int) SingleNodeMiningDistribution {
	var pathToNodeLogs = getPathToNodeLogs(folderWithNodesLogs, nodeNumber)
	var nodeEvents = NodeEvents.LoadNodeEvents(pathToNodeLogs)

	var nodeBlockTree = getNodeBlockTreeFromEvents(nodeEvents)
	var nodeMainChain = getMainChain(nodeBlockTree)
	var minerDistribution = getMinerDistribution(nodeMainChain)

	var minedByOwn = minerDistribution[0]
	var totalBlocks = minerDistribution[0] + minerDistribution[-1]

	return SingleNodeMiningDistribution {minedByOwn, totalBlocks}
}

func getNodeBlockTreeFromEvents(nodeEvents []NodeEvents.NodeEvent) map[string]BlockInfo {
	const minedByOwn = 0
	const minedByOthers = -1
	blocksByHash := make(map[string]BlockInfo)

	// Get the initial map of blocks by hash (without height)
	for _, nodeEvent := range nodeEvents {
		switch event := nodeEvent.Event.(type) {
		case NodeEvents.MinedBlockEvent:
			var blockInfo BlockInfo
			blockInfo.BlockHash = event.BlockHash
			blockInfo.ParentBlock = event.ParentBlock
			blockInfo.MinerNode = minedByOwn
			blockInfo.Height = -1

			blocksByHash[event.BlockHash] = blockInfo
		case NodeEvents.SelfishMinedBlockEvent:
			var blockInfo BlockInfo
			blockInfo.BlockHash = event.BlockHash
			blockInfo.ParentBlock = event.ParentBlock
			blockInfo.MinerNode = minedByOwn
			blockInfo.Height = -1

			blocksByHash[event.BlockHash] = blockInfo
		case NodeEvents.BroadcastedBlockEvent:
			var blockInfo BlockInfo
			blockInfo.BlockHash = event.BlockHash
			blockInfo.ParentBlock = event.ParentBlock
			blockInfo.MinerNode = minedByOthers
			blockInfo.Height = -1

			blocksByHash[event.BlockHash] = blockInfo
		}
	}	

	// Calculate the height of each block
	for blockHash, _ := range blocksByHash {
		updateHeightOfBlockAndAncestors(blockHash, blocksByHash)
	}
	return blocksByHash
}


/** Common helpers **/
// Updates the block height of the passed block and all it's ancestors
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

// Given a block, returns it and all its ancestors
func allAncestorsOf(blockHash string, allBlocks map[string]BlockInfo) []BlockInfo {
	var blockInfo = allBlocks[blockHash]
	var parentBlockHash = blockInfo.ParentBlock
	if _, isParentBlockContained := allBlocks[parentBlockHash]; isParentBlockContained {
		var ancestorsOfParent = allAncestorsOf(parentBlockHash, allBlocks)
		return append(ancestorsOfParent, blockInfo)
	} else {
		return append([]BlockInfo{}, blockInfo)
	}
}

// Given a block tree of blocks, obtains the main chain
func getMainChain(blocksByHash map[string]BlockInfo) []BlockInfo {
	// Obtain the block with the biggest height
	var highestHeight = -1
	var blockWithHighestHeight = ""
	for blockHash, blockInfo := range blocksByHash {
		if blockInfo.Height > highestHeight {
			highestHeight = blockInfo.Height
			blockWithHighestHeight = blockHash
		}
	}

	if (blockWithHighestHeight == "") {
		return nil
	} else {
		return allAncestorsOf(blockWithHighestHeight, blocksByHash)
	}
}

func getMinerDistribution(mainChain []BlockInfo) map[int]int {
	minerDistribution := make(map[int]int)

	for _, blockInfo := range mainChain {
		minerDistribution[blockInfo.MinerNode] = minerDistribution[blockInfo.MinerNode] + 1
	}
	return minerDistribution
}

func getPathToNodeLogs(folderWithNodesLogs string, nodeNumber int) string {
	var nodeNumberAsString = strconv.Itoa(nodeNumber)
	var pathToNodeLogs = folderWithNodesLogs + "/btcCoreLogN" + nodeNumberAsString

	return pathToNodeLogs
}