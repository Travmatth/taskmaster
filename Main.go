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
	config := os.Args[1]
	logOut := os.Args[2]
	debugLevel := os.Args[3]
	if jobs, err := LoadJobsFromFile(config); err != nil {
		fmt.Println(err)
	} else if logErr := NewLogger(logOut, debugLevel); logErr != nil {
		fmt.Println(logErr)
	} else {
		s := NewSupervisor(config, NewManager())
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-c
			if sig == syscall.SIGHUP {
				if reloadJobs, err := LoadJobsFromFile(config); err != nil {
					Log.Info("Error reloading configuration", err)
					s.StopAllJobs(true)
					os.Exit(1)
				} else {
					Log.Info("Supervisor: signal", sig, "received, reloading", config)
					s.Reload(reloadJobs, false)
				}
			} else {
				s.StopAllJobs(true)
				os.Exit(0)
			}
		}()
		for {
			if err := s.Reload(jobs, false); err != nil {
				panic(err)
			} else {
				for {
				}
			}
		}
	}
}
