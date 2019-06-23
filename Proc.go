package main

import (
	"errors"
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

// Restart will kill the currently running process with the given Pid,
// and start again with new attributes
func (p *Proc) Restart(pid int, pids map[int]*os.Process) (int, error) {
	return 0, nil
}

// PerformStartCheckup called after process started, moves status: PROCSTART -> PROCRUNNING
func (p *Proc) PerformStartCheckup() {
	timeout := time.Duration(p.StartCheckup) * time.Second
	time.Sleep(timeout)
	p.status = PROCRUNNING
	p.start = time.Now()
}

// WaitForExit catches process exit, removes process pids map, moves status: PROCRUNNING -> PROCSTOPPED
func (p *Proc) WaitForExit(pids map[int]*os.Process, process *os.Process) {
	_, _ = process.Wait()
	p.status = PROCSTOPPED
	delete(pids, p.pid)
}

// Start will exec the process it is called on
func (p *Proc) Start(pids map[int]*os.Process) (int, error) {
	if p.status != PROCSTOPPED {
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
		return 0, err
	}
	p.status = PROCRUNNING
	p.pid = process.Pid
	pids[p.pid] = process
	syscall.Umask(defaultUmask)
	go p.PerformStartCheckup()
	go p.WaitForExit(pids, process)
	return p.pid, nil
}
