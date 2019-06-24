package main

import (
	"fmt"
	"os"
)

func main() {
	var supervisor Supervisor
	m := NewManager(&supervisor, make(chan ProcEvent))

	if l := len(os.Args); l < 3 {
		fmt.Println("Usage: ./taskmaster <Config_File> <Log_File> [Log_level]\n")
	} else if configProcs, err := LoadConfig(os.Args[1:]); err == nil {
		newProcs := SetDefaults(configProcs)
		supervisor.procs = make(map[int]*Proc)
		supervisor.processes = make(map[int]*os.Process)
		go m.ManageProcs()
		supervisor.TestAll(newProcs, m.event)
	} else {
		fmt.Println(err)
	}
}
