package main

import (
	"fmt"
	"os"
)

func main() {
	var supervisor Supervisor

	if l := len(os.Args); l != 2 {
		fmt.Print("Error: Please specify a configuration file\n")
	} else {
		yamlFile := os.Args[1]
		configProcs := LoadConfig(yamlFile)
		newProcs := SetDefaults(configProcs)
		supervisor.processes = make(map[int]*os.Process)
		supervisor.TestAll(newProcs)
	}
}
