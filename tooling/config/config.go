package config

type Configuration struct {
	// Location of the needed scripts
	NodeExecutionDir    string
	AddressesDir        string
	SherlockfogDir      string
	TopologyCreationDir string

	// Constants
	BlockIntervalInSeconds float64
}

// FIXME: Maybe this can be derived from the current location of this file?
// Allows runnning all the scripts using this from anywhere
// var projectToolingRoot = "/Users/mpizzagali/Tesis/btc-core/tooling"
var projectToolingRoot = "/home/mgeier/mpizzagalli/bitcoin/tooling"
var config = Configuration{
	NodeExecutionDir:    projectToolingRoot + "/ejecucion-nodos",
	AddressesDir:        projectToolingRoot + "/ejecucion-nodos/.exec-results",
	SherlockfogDir:      projectToolingRoot + "/sherlockfog",
	TopologyCreationDir: projectToolingRoot + "/creacion-topologia",

	BlockIntervalInSeconds: 10.0,
}

// Usa PWD y la estructura del proyecto para obtener varios paths a ser utilizados por otros archivos
func GetConfiguration() Configuration {
	return config
}
