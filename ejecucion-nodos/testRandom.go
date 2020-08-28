package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	Config "../config"
	Utils "../utils"
)

// FIXME: Capaz convendria comentar los logs... total esta info no va a ser usada para nada
// FIXME: Hay un archivo de log pero que no se crea... que onda eso?
// FIXME: Es una paja que tarde tanto en generar las cosas, hay cosas que no entiendo porque, pero las otras, deberia ser por config, tener dos versiones, por un flag de debug o que onda?

var config = Config.GetConfiguration()
var blockIntervalInSeconds = 2.0

// Mina bloques de acuerdo al hashing power asignado (simuLambda)
func mineBlocks() {

	nodeNumber := os.Args[1]
	simuLambda, _ := strconv.ParseFloat(os.Args[2], 64)

	var sleepTime float64
	var sleepSeconds time.Duration
	var sleepNanoseconds time.Duration
	var nextBlockTime time.Duration

	rng := Utils.CreateRng()

	timestamp := time.Now().UnixNano()

	for i := 0; ; i ^= 1 {
		// Hardcodeado que el tiempo entre bloques sea propocional a los 10 min
		random := rng.ExpFloat64()
		sleepTime = (random / simuLambda) * blockIntervalInSeconds
		sleepSeconds = time.Duration(sleepTime)
		sleepNanoseconds = time.Duration((sleepTime - float64(sleepSeconds)) * 1000000000.0)
		nextBlockTime = time.Duration(timestamp) + sleepSeconds*time.Second + sleepNanoseconds

		time.Sleep(nextBlockTime - time.Duration(time.Now().UnixNano()))

		timestamp = time.Now().UnixNano()

		// fmt.Println(time.Now())
		fmt.Println(time.Now(), " [testEngine] ", nodeNumber, ": Mining a block!")
		// os.Stdout.WriteString(fmt.Sprintf("[testEngine] %s: Mining a block!\n", nodeNumber))
	}
}

/*
	Para levantar mas rapidamente:
	- Comentar el tiempo de espera inicial al cargar las addresses
	- Disminuir el tiempo entre bloques de 10min a algo corto
	- Disminuir el tiempo hasta que se empieza a generar la primer tx
*/
// Manda txs al nodo y mina bloques incluyendolas, hasta que se mata al nodo con una señal de sigterm
func main() {
	fmt.Println("Starting to mine with params:")
	fmt.Println(os.Args)
	mineBlocks()
	fmt.Println("Already sent to mine")
}
