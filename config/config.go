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
var projectRoot = "/Users/mpizzagali/Tesis/btc-core"
var config = Configuration{
	NodeExecutionDir:    projectRoot + "/ejecucion-nodos",
	AddressesDir:        projectRoot + "/ejecucion-nodos/.exec-results",
	SherlockfogDir:      projectRoot + "/sherlockfog",
	TopologyCreationDir: projectRoot + "/creacion-topologia",

	BlockIntervalInSeconds: 600.0,
}

// Usa PWD y la estructura del proyecto para obtener varios paths a ser utilizados por otros archivos
func GetConfiguration() Configuration {
	return config
}
