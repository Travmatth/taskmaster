package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
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
func (p *Proc) Kill(s *Supervisor) error {
	fmt.Println("Killing proc: ", p.ID)
	if p.status != PROCSTOPPED {
		fmt.Println("Killing Proces: ", p.pid, ", ", s.processes[p.pid].Pid)
		err := s.processes[p.pid].Kill()
		delete(s.processes, p.pid)
		p.status = PROCSTOPPED
		fmt.Println("Killed proc: ", p.ID)
		return err
	}
	fmt.Println("Killed proc: ", p.ID, " already stopped")
	return nil
}

// Start will exec the process it is called on
func (p *Proc) Start(s *Supervisor) (int, error) {
	fmt.Println("Starting proc: ", p.ID)
	if p.status != PROCSTOPPED {
		fmt.Println("Start proc: ", p.ID, " already started")
		return 0, errors.New("Error: Process not currently stopped")
	}
	defaultUmask := syscall.Umask(p.Umask)
	process, err := os.StartProcess(p.Args[0], p.Args, &os.ProcAttr{
		Dir:   p.WorkingDir,
		Env:   p.EnvVars,
		Files: p.Redirections,
	})
	if err != nil {
		p.status = PROCSTOPPED
		fmt.Println("Start proc: ", p.ID, " encountered error in os.StartProcess")
		return 0, err
	}
	p.status = PROCRUNNING
	p.pid = process.Pid
	s.processes[p.pid] = process
	syscall.Umask(defaultUmask)
	go p.PerformStartCheckup(s, process)
	fmt.Println("Started proc: ", p.ID)
	return p.pid, nil
}

// Restart will kill the currently running process with the given Pid,
// and start again with new attributes
func (p *Proc) Restart(s *Supervisor) (int, error) {
	fmt.Println("Restarting proc: ", p.ID)
	if err := p.Kill(s); err != nil {
		return 0, err
	}
	pid, err := p.Start(s)
	fmt.Println("Restarted proc: ", p.ID)
	return pid, err
}
