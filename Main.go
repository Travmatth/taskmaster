package main

import (
	"fmt"
	"os"
	"syscall"

	. "github.com/Travmatth/taskmaster/log"
	PARSE "github.com/Travmatth/taskmaster/parse"
	SIG "github.com/Travmatth/taskmaster/signals"
	SVSR "github.com/Travmatth/taskmaster/supervisor"
	UI "github.com/Travmatth/taskmaster/ui"
)

type Opts struct {
	Config string
	Log    string
	Level  string
}

func parseOpts(args []string) (opts Opts, ok bool) {
	ok = true
	if len(args) == 3 {
		opts.Level = "4"
	} else if len(args) == 4 {
		opts.Level = args[3]
	} else {
		ok = false
		return
	}
	opts.Config, opts.Log = args[1], args[2]
	return
}

//ManageSignals handles the responses to signals sent to the program
func ManageSignals(s *SVSR.Supervisor, config string, c chan os.Signal) {
	sig := <-c
	if sig == syscall.SIGHUP {
		if reloadJobs, err := PARSE.LoadJobsFromFile(config); err != nil {
			Log.Info("Error reloading configuration", err)
			s.StopAllJobs(true)
			os.Exit(1)
		} else {
			Log.Info("Supervisor: signal", sig, "received, reloading", config)
			s.Reload(reloadJobs, false)
		}
	} else if sig == syscall.SIGTERM || sig == syscall.SIGINT {
		Log.Info("Supervisor: exit signal received, shutting down")
		s.StopAllJobs(true)
		os.Exit(0)
	}
	go ManageSignals(s, config, c)
}

func main() {
	if opts, ok := parseOpts(os.Args); ok == false {
		fmt.Println("Usage: ./taskmaster <Config_File> <Log_File> [Log_Level]")
		fmt.Println("\tConfig_File: Procfile you wish to run")
		fmt.Println("\tLog_File: Log file you wish to use")
		levels := "0 CRITICAL, 1 ERROR, 2 WARNING, 3 NOTICE, 4 INFO, 5 DEBUG"
		fmt.Println("\tLog_Level: ", levels)
	} else if jobs, err := PARSE.LoadJobsFromFile(opts.Config); err != nil {
		fmt.Println(err)
	} else if err := NewLogger(opts.Log, opts.Level); err != nil {
		fmt.Println(err)
	} else {
		s := SVSR.NewSupervisor(opts.Config, opts.Log,
			SVSR.NewManager(), SIG.InitSignals())
		go ManageSignals(s, opts.Config, s.SigCh)
		for {
			if err := s.Reload(jobs, false); err != nil {
				Log.Info("Error reloading configuration", err)
				s.StopAllJobs(true)
				os.Exit(1)
			}
			f := UI.NewFrontend(s)
			f.StartUI()
		}
	}
}
