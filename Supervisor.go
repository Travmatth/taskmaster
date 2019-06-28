package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

// Supervisor struct contains array of Proc structs
type Supervisor struct {
	procs     map[int]*Proc
	processes map[int]*os.Process
	mutex     sync.Mutex
}

// NewSupervisor creates a Supervisor struct
func NewSupervisor() *Supervisor {
	var supervisor Supervisor
	supervisor.procs = make(map[int]*Proc)
	supervisor.processes = make(map[int]*os.Process)
	return &supervisor
}

// Kill will immediately exit the given process if running, silently fail otherwise
func (s *Supervisor) Kill(p *Proc) error {
	if p.Status != PROCSTOPPED {
		p.Status = PROCSTOPPED
		err := s.processes[p.pid].Kill()
		if err == nil {
			s.SafeDeleteProc(p)
			s.SafeDeletePid(p)
			Log.Info("Process ", p.ID, " killed")
		}
		return err
	}
	Log.Info("Process ", p.ID, " not running")
	return nil
}

// Start will exec the process it is called on
func (s *Supervisor) Start(p *Proc) error {
	if p.Status != PROCSTOPPED {
		return errors.New("Error: Process not currently stopped")
	}
	defaultUmask := syscall.Umask(p.Umask)
	process, err := os.StartProcess(p.Args[0], p.Args, &os.ProcAttr{
		Dir:   p.WorkingDir,
		Env:   p.EnvVars,
		Files: p.Redirections,
	})
	if err != nil {
		Log.Debug("Start err setting status to STOPPROC")
		p.Status = PROCSTOPPED
		return err
	}
	p.Status = PROCSTART
	p.pid = process.Pid
	s.SafeAddProc(p)
	s.SafeAddPid(p, process)
	syscall.Umask(defaultUmask)
	// go p.PerformStartCheckup(m.event, process)
	go func() {
		time.Sleep(time.Duration(p.StartCheckup) * time.Second)
		m.channels.setrunning <- p
	}()
	go func() {
		state, err := process.Wait()
		Log.Debug("WaitForExit wait reached, exit code: ", state.ExitCode(), " error: ", err)
		m.channels.exited <- p
		// m.channels.setstopped <- p
		if err != nil {
			Log.Info("Error encountered waiting for exit: ", state.ExitCode())
			return
		}
	}()
	return nil
}

// Stop sends the specified stop signal to the process
func (s *Supervisor) Stop(p *Proc) error {
	if p.Status == PROCSTOPPED {
		return fmt.Errorf("Process %d already stopped", p.ID)
	}
	// go p.StopTimeoutCheck(m.event)
	go func() {
		Log.Debug("Checking Process ", p.ID, " exit timeout, status: ", p.Status)
		select {
		case <-time.After(time.Duration(p.StopTimeout) * time.Second):
			Log.Info("Error: Process ", p.ID, " did not exit gracefully, killing")
			m.channels.kill <- p
		case p := <-m.channels.exited:
			Log.Debug("Process ", p.ID, " is stopped on timeout check")
			m.channels.setstopped <- p
		}
	}()
	Log.Debug("Sending ", p.StopSignal, " to Process ", p.ID)
	m.supervisor.processes[p.pid].Signal(p.StopSignal)
	return nil
}

// Restart will kill the currently running process with the given Pid,
// and start again with new attributes
func (s *Supervisor) Restart(p *Proc) error {
	if err := m.Stop(p); err != nil {
		Log.Info("Error stopping Process ID: ", p.ID, " ", err)
		if err = m.Kill(p); err != nil {
			return err
		}
		Log.Info("Process ID: ", p.ID, " not responding to stop signal, killed")
	} else {
		Log.Info("Process ID: ", p.ID, " Stopped")
	}
	if err := m.Start(p); err != nil {
		return err
	}
	Log.Info("Process ID: ", p.ID, " Started")
	return nil
}

// StartAll iterates through given processes, launching each
func (s *Supervisor) StartAll(configProcs []Proc, isLaunch bool) {
	newProcs, changedProcs := s.DiffProcs(configProcs)
	for _, proc := range newProcs {
		if isLaunch && proc.AtLaunch {
			proc.Start(s.channels)
		} else if !isLaunch {
			proc.Start(s.channels)
		}
	}
	for _, proc := range changedProcs {
		proc.Restart(s.channels)
	}
}

// TestAll launches proc test - starts all proceses, then kills all processes
func (s *Supervisor) TestAll(configProcs []Proc) {
	s.StartAll(configProcs, true)
	time.Sleep(time.Duration(5) * time.Second)
	// for _, p := range s.procs {
	// 	p.Restart(events)
	// }
	for _, p := range s.procs {
		p.Stop(s.channels)
	}
	time.Sleep(time.Duration(10) * time.Second)
}
