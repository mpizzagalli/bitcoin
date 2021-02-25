package main

import (
	"os"
	"fmt"
	"strconv"
	Parsing "./parsing"
)

func main(){
	var command = os.Args[1]

	switch command {
	case "obtained-single":
		/*
			Parameters:
			- Node number
			- Folder with logs of the nodes (btcCoreLogN_)

			Prints the miner distribution of the main chain produced by the nodes
 		*/
		var nodeNumber, _ = strconv.Atoi(os.Args[2])
		var folderWithLogs = os.Args[3]
	
		var singleMiningDistribution = Parsing.CalculateMiningDistributionFromNodePerspective(folderWithLogs, nodeNumber)
		os.Stdout.WriteString(fmt.Sprintf("Own: %d/total: %d\n", singleMiningDistribution.Own, singleMiningDistribution.Total))
	case "obtained-all":
		/*
			Parameters:
			- Number of nodes
			- Folder with logs of the nodes (btcCoreLogN_)

			Prints the miner distribution of the main chain produced by the nodes
 		*/
		var numberOfNodes, _ = strconv.Atoi(os.Args[2])
		var folderWithLogs = os.Args[3]
	
		Parsing.CalculateAndLogMiningDistribution(numberOfNodes, folderWithLogs)
	case "obtained-intervals":
		var numberOfNodes, _ = strconv.Atoi(os.Args[2])
		var folderWithLogs = os.Args[3]
	
		Parsing.CalculateMiningDistributionInIntervals(numberOfNodes, folderWithLogs, 25)
	case "expected":
		/*
			Parameters:
			- alpha
			- delta

			Prints the expected percentage of blocks in the main chain mined by the selfish miner
 		*/
		var alpha, _ = strconv.ParseFloat(os.Args[2], 64)
		var delta, _ = strconv.ParseFloat(os.Args[3], 64)

		var expectedSelfishMiningProportion = Parsing.CalculateExpectedSelfishMiningDistribution(alpha, delta)

		os.Stdout.WriteString(fmt.Sprintf("Selfish should mine %f\n", expectedSelfishMiningProportion))
	case "experiments-all":
		var folderWithResults = os.Args[2]
		Parsing.CalculateRewardDistributionExperiments(folderWithResults)
	case "experiments-selfish":
		var folderWithResults = os.Args[2]
		Parsing.CalculateSelfishMinerProportionExperiments(folderWithResults)
	}
}