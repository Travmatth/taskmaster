package main

import (
	"fmt"
	"os"
)

func run(s *Supervisor, newJobs []*Job) {
	for {
		if err := s.Reload(newJobs); err != nil {
			panic(err)
		} else {
			s.WaitForExit()
		}
	}
}

func main() {
	if l := len(os.Args); l < 3 {
		fmt.Println("Usage: ./taskmaster <Config_File> <Log_File> [Log_level]")
	} else if buf, fileErr := LoadFile(os.Args[1]); fileErr != nil {
		fmt.Println(fileErr)
	} else if jobConfigs, configErr := LoadJobs(buf); configErr != nil {
		fmt.Println(configErr)
	} else if Log, logErr := NewLogger(os.Args[1:]); configErr != nil {
		fmt.Println(logErr)
	} else {
		s := NewSupervisor(os.Args[1], Log, NewManager(Log))
		newJobs := SetDefaults(jobConfigs, s.Log)
		run(s, newJobs)
		s.Log.Info("Exiting")
	}
}
