package main

import (
	"fmt"
	"os"
	"time"
)

// Proc structs contain basic information of launched processes
type Proc struct {
	Command       string
	Instances     int
	AtLaunch      bool
	RestartPolicy string
	ExpectedExit  string
	StartCheckup  string
	MaxRestarts   string
	StopSignal    string
	KillTimeout   int
	Redirections  struct {
		Stdout string
		Stderr string
	}
	EnvVars    string
	WorkingDir string
	Umask      string
	start      time.Time
	Process    os.Process
}

// Supervisor struct contains array of Proc structs
type Supervisor struct {
	procs []Proc
}

func (s *Supervisor) Start() {
	fmt.Print("Start")
}
