package main

import (
	"errors"
	"os"
	"syscall"
	"time"
)

const (
	// ProcStopped signifies process not currently running
	ProcStopped = iota
	// ProcRunning signifies process currently running
	ProcRunning = iota
	// ProcStart signifies process beginning its start sequence
	ProcStart = iota
)

const (
	// RestartAlways signifies restart whenever process is stopped
	RestartAlways = iota
	// RestartNever signifies never restart process
	RestartNever = iota
	// RestartUnexpected signifies restart on unexpected exit
	RestartUnexpected = iota
)

// Proc structs contain basic information of launched processes
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
	Redirections  struct {
		Stdin  *os.File
		Stdout *os.File
		Stderr *os.File
	}
}

// Restart will kill the currently running process with the given Pid,
// and start again with new attributes
func (p *Proc) Restart(pid int, pids map[int]os.Process) (int, error) {
	var newPid int
	return newPid, nil
}

// Start will exec the process it is called on
func (p *Proc) Start(pids map[int]os.Process) (int, error) {
	var newPid int
	var defaultUmask int

	if p.status != ProcStopped {
		return 0, errors.New("Error: Process not currently stopped")
	}
	defaultUmask = syscall.Umask(p.Umask)
	// use os.Command instead?
	// process, err := os.StartProcess(p.Args[0], p.Args, &os.ProcAttr{
	// 	Dir:   p.WorkingDir,
	// 	Env:   strings.Fields(p.EnvVars),
	// 	Files: p.Redirections,
	// })
	syscall.Umask(defaultUmask)
	return newPid, nil
}
