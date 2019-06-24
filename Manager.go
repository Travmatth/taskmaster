package main

import (
	"errors"
	"os"
	"syscall"
	"time"
)

// Enums to signal action to take on Proc
const (
	KILLPROC       = iota
	STOPPROC       = iota
	STARTPROC      = iota
	RESTARTPROC    = iota
	SETPROCRUNNING = iota
	SETPROCSTOPPED = iota
)

// ProcEvent sent over channels to manage *Proc's
type ProcEvent struct {
	Action int
	proc   *Proc
}

// Manager handles channel messages
type Manager struct {
	supervisor *Supervisor
	event      chan ProcEvent
}

// NewManager inits and returns the Manager struct
func NewManager(supervisor *Supervisor, events chan ProcEvent) *Manager {
	var manager Manager
	manager.supervisor = supervisor
	manager.event = events
	return &manager
}

// Kill will immediately exit the given process if running, silently fail otherwise
func (m *Manager) Kill(p *Proc) error {
	Log.Debug("manager.Kill() start: ", p.ID)
	if p.status != PROCSTOPPED {
		Log.Debug("Killing Proces: ", p.pid, ", ", m.supervisor.processes[p.pid].Pid)
		p.status = PROCSTOPPED
		err := m.supervisor.processes[p.pid].Kill()
		delete(m.supervisor.procs, p.ID)
		delete(m.supervisor.processes, p.pid)
		Log.Info("Killed proc: ", p.ID)
		return err
	}
	Log.Info("Killed proc: ", p.ID, " already stopped")
	return nil
}

// Start will exec the process it is called on
func (m *Manager) Start(p *Proc) error {
	Log.Debug("Starting proc: ", p.ID)
	if p.status != PROCSTOPPED {
		Log.Info("Start proc: ", p.ID, " already started")
		return errors.New("Error: Process not currently stopped")
	}
	defaultUmask := syscall.Umask(p.Umask)
	process, err := os.StartProcess(p.Args[0], p.Args, &os.ProcAttr{
		Dir:   p.WorkingDir,
		Env:   p.EnvVars,
		Files: p.Redirections,
	})
	if err != nil {
		p.status = PROCSTOPPED
		Log.Info("Start proc: ", p.ID, " encountered error in os.StartProcess")
		return err
	}
	p.status = PROCRUNNING
	p.pid = process.Pid
	m.supervisor.procs[p.ID] = p
	m.supervisor.processes[p.pid] = process
	syscall.Umask(defaultUmask)
	go p.PerformStartCheckup(m.event, process)
	Log.Info("Started proc: ", p.ID)
	return nil
}

// Restart will kill the currently running process with the given Pid,
// and start again with new attributes
func (m *Manager) Restart(p *Proc) error {
	Log.Debug("Restarting proc: ", p.ID)
	if err := m.Kill(p); err != nil {
		return err
	} else if err = m.Start(p); err != nil {
		return err
	}
	Log.Info("Restarted proc: ", p.ID)
	return nil
}

// ManageProcs waits for actions to be taken on Proc's, executes them
func (m *Manager) ManageProcs() {
	for {
		msg := <-m.event
		p := msg.proc
		switch msg.Action {
		case STOPPROC:
			Log.Debug("STOPPROC start")
			Log.Debug("STOPPROC end")
			fallthrough
		case KILLPROC:
			Log.Debug("KILLPROC start")
			err := m.Kill(p)
			if err != nil {
				Log.Debug(err)
			}
			Log.Debug("KILLPROC end")
		case RESTARTPROC:
			Log.Debug("RESTARTPROC start")
			err := m.Kill(p)
			if err != nil {
				Log.Debug(err)
				break
			}
			Log.Debug("RESTARTPROC end")
			fallthrough
		case STARTPROC:
			Log.Debug("STARTPROC start")
			err := m.Start(p)
			if err != nil {
				Log.Debug(err)
			}
			Log.Debug("STARTPROC end")
		case SETPROCRUNNING:
			Log.Debug("PerformStartCheckup ", p.ID, " is running ", p.status)
			p.status = PROCRUNNING
			p.start = time.Now()
		case SETPROCSTOPPED:
			if p.status == PROCRUNNING {
				p.status = PROCSTOPPED
				delete(m.supervisor.processes, p.pid)
				Log.Debug("WaitForExit ", p.ID, " has stopped ", p.status)
			}
		}
	}
}
