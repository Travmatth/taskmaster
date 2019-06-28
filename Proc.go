package main

import (
	"os"
	"time"
)

const (
	// PROCSTOPPED signifies process not currently running
	PROCSTOPPED = iota
	// PROCRUNNING signifies process currently running
	PROCRUNNING = iota
	// PROCSTART signifies process beginning its start sequence
	PROCSTART = iota
)

const (
	// RESTARTALWAYS signifies restart whenever process is stopped
	RESTARTALWAYS = iota
	// RESTARTNEVER signifies never restart process
	RESTARTNEVER = iota
	// RESTARTUNEXPECTED signifies restart on unexpected exit
	RESTARTUNEXPECTED = iota
)

// Proc structs track information of processes given in configuration
type Proc struct {
	ID            int
	Args          []string
	Instances     int
	AtLaunch      bool
	RestartPolicy int
	ExpectedExit  os.Signal
	StartCheckup  int
	MaxRestarts   int
	StopSignal    os.Signal
	StopTimeout   int
	EnvVars       []string
	WorkingDir    string
	Umask         int
	start         time.Time
	pid           int
	Status        int
	Redirections  []*os.File
	end           chan bool
}
