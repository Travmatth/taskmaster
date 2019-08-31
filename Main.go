package main

import (
	"fmt"
	"os"
)

func main() {
	if l := len(os.Args); l < 3 {
		fmt.Println("Usage: ./taskmaster <Config_File> <Log_File> [Log_level]")
		return
	}
	config, logOut, debugLevel := os.Args[1], os.Args[2], os.Args[3]
	if jobs, err := LoadJobsFromFile(config); err != nil {
		fmt.Println(err)
	} else if logErr := NewLogger(logOut, debugLevel); logErr != nil {
		fmt.Println(logErr)
	} else {
		s := NewSupervisor(config, NewManager(), InitSignals(), logOut)
		go ManageSignals(s, config, s.SigCh)
		for {
			if err := s.Reload(jobs, false); err != nil {
				Log.Info("Error reloading configuration", err)
				s.StopAllJobs(true)
				os.Exit(1)
			}
			StartUI(s)
		}
	}
}
