package main

import (
	"fmt"
	"os"
)

func main() {
	// var supervisor Supervisor

	if l := len(os.Args); l != 2 {
		fmt.Print("Error: Please specify a configuration file\n")
	} else {
		yamlFile := os.Args[1]
		configProcs := LoadConfig(yamlFile)
		fmt.Println(configProcs)
		newProcs := SetDefaults(configProcs)
		fmt.Println(newProcs)
		// supervisor.StartAll(newProcs)
	}
}
