package main

import (
	"fmt"
	"os"
)

func main() {
	if l := len(os.Args); l < 3 {
		fmt.Println("Usage: ./taskmaster <Config_File> <Log_File> [Log_level]")
	} else if buf, fileErr := LoadFile(os.Args[1]); fileErr != nil {
		fmt.Println(fileErr)
	} else if configProcs, configErr := LoadJobs(buf); configErr != nil {
		fmt.Println(configErr)
	} else if Log, logErr := NewLogger(os.Args[1:]); configErr != nil {
		fmt.Println(logErr)
	} else {
		s := NewSupervisor(Log)
		newJobs := SetDefaults(configProcs)
		s.StartAllJobs(newJobs)
		s.Log.Info("Exiting")
	}
}
