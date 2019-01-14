package main

import (
	"math/rand"
 	"strconv"
 	"os"
	"time"
	"fmt"
)

func runExperiment(n int) map[int]int {

	rndSlice := []int{0,0,0,0,0,0}

	nodeRndGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	grados := make(map[int]int)

	for i:=1; i<n; i++ {

		r := nodeRndGen.Int63n(int64(len(rndSlice)))

		nodeA := rndSlice[r]

		grados[i]++
		grados[nodeA]++

		rndSlice = append(rndSlice[:r], rndSlice[r+1:]...)

		r = nodeRndGen.Int63n(int64(len(rndSlice)))

		nodeB := rndSlice[r]

		if (nodeA != nodeB) {
			rndSlice = append(rndSlice[:r], rndSlice[r+1:]...)
			grados[i]++
			grados[nodeB]++
		}

		/*r = nodeRndGen.Int63n(int64(len(rndSlice)))

		nodeC := rndSlice[r]

		if (nodeA != nodeC && nodeB != nodeC) {
			rndSlice = append(rndSlice[:r], rndSlice[r+1:]...)
			grados[i]++
			grados[nodeC]++
		}*/

		rndSlice = append(rndSlice, []int{i,i,i,i,i,i}...)

	}

	stats := make(map[int]int)

	/*for _, v := range grados {
		//fmt.Println(fmt.Sprintf("EL Nodo %d tiene grado %d", k, v))

		stats[v]++

		/*if (v>6) {
			//fmt.Println(fmt.Sprintf("EL Nodo %d tiene grado %d", k, v))
		} else {
			fmt.Println(v)
		}*/
	//}

	for i:=0; i<n; i++ {
		stats[grados[i]]++
	}

	return stats
}

func main() {

	n , _:= strconv.Atoi(os.Args[1])

	k , _:= strconv.Atoi(os.Args[2])

	v := make([]map[int]int, 0, k)

	for i:=0; i<k; i++ {
		v = append(v, runExperiment(n))
	}

	d := float64(k)

	tot := 0.0;

	m := float64(n)

	for i:=0; i<11; i++ {

		sum := 0

		for j:=0; j<k; j++ {
			sum += v[j][i]
		}

		avg := float64(sum)/d

		fmt.Println(fmt.Sprintf("%d %f", i, (avg/m)*100.0))

		tot += avg*float64(i)

	}
	fmt.Println(fmt.Sprintf("total aristas %f", tot))

}
