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
	KillTimeout   int
	EnvVars       []string
	WorkingDir    string
	Umask         int
	start         time.Time
	pid           int
	status        int
	Redirections  []*os.File
}

// Kill will immediately exit the given process if running, silently fail otherwise
func (p *Proc) Kill(events chan ProcEvent) {
	events <- ProcEvent{KILLPROC, p}
}

// Start will exec the process it is called on
func (p *Proc) Start(events chan ProcEvent) {
	events <- ProcEvent{STARTPROC, p}
}

// Restart will kill the currently running process with the given Pid,
// and start again with new attributes
func (p *Proc) Restart(events chan ProcEvent) {
	events <- ProcEvent{RESTARTPROC, p}
}
