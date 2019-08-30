package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if l := len(os.Args); l < 3 {
		fmt.Println("Usage: ./taskmaster <Config_File> <Log_File> [Log_level]")
		return
	}
	configFile := os.Args[1]
	if jobs, err := LoadJobsFromFile(configFile); err != nil {
		fmt.Println(err)
	} else if logErr := NewLogger(os.Args[1:]); logErr != nil {
		fmt.Println(logErr)
	} else {
		s := NewSupervisor(configFile, NewManager())
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-c
			if sig == syscall.SIGHUP {
				if reloadJobs, err := LoadJobsFromFile(configFile); err != nil {
					Log.Info("Error reloading configuration", err)
					s.StopAllJobs()
					os.Exit(1)
				} else {
					Log.Info("Supervisor: signal", sig, "received, reloading", configFile)
					s.Reload(reloadJobs)
				}
			} else {
				s.StopAllJobs()
				os.Exit(0)
			}
		}()
		for {
			if err := s.Reload(jobs); err != nil {
				panic(err)
			} else {
				for {
				}
			}
		}
	}
}
