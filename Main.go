package main

import (
	"fmt"
	"os"
)

func main() {
	s := NewSupervisor()
	if l := len(os.Args); l < 3 {
		fmt.Println("Usage: ./taskmaster <Config_File> <Log_File> [Log_level]")
	} else if configProcs, err := LoadConfig(os.Args[1:]); err == nil {
		newProcs := SetDefaults(configProcs)
		s.TestAll(newProcs)
	} else {
		fmt.Println(err)
	}
	Log.Info("Exiting")
}
