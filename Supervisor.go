package main

import (
	"os"
	"time"
)

// Supervisor struct contains array of Proc structs
type Supervisor struct {
	procs     map[int]*Proc
	processes map[int]*os.Process
}

/*
DiffProcs identifies which processes are new and should be started,
which processes are changed and should be stopped and restarted,
and which processes are unchanged
*/
func (s *Supervisor) DiffProcs(configProcs []Proc) ([]Proc, []Proc) {
	Log.Debug("DiffProcs Start")
	newProcs := []Proc{}
	changedProcs := []Proc{}

	for _, proc := range configProcs {
		_, ok := s.procs[proc.ID]
		if ok == false {
			Log.Debug("New proc detected")
			newProcs = append(newProcs, proc)
		} else {
			Log.Debug("Changed proc detected")
			changedProcs = append(changedProcs, proc)
		}
	}
	Log.Debug("DiffProcs End")
	return newProcs, changedProcs
}

// StartAll iterates through given processes, launching each
func (s *Supervisor) StartAll(configProcs []Proc, events chan ProcEvent, isLaunch bool) {
	Log.Debug("StartAll Start")
	newProcs, changedProcs := s.DiffProcs(configProcs)
	for _, proc := range newProcs {
		if isLaunch && proc.AtLaunch {
			proc.Start(events)
		} else if !isLaunch {
			proc.Start(events)
		}
	}
	for _, proc := range changedProcs {
		proc.Restart(events)
	}
	Log.Debug("StartAll End")
}

// TestAll launches proc test - starts all proceses, then kills all processes
func (s *Supervisor) TestAll(configProcs []Proc, events chan ProcEvent) {
	Log.Debug("TestAll Start")
	s.StartAll(configProcs, events, true)
	time.Sleep(time.Duration(5) * time.Second)
	for _, p := range s.procs {
		p.Kill(events)
	}
	Log.Debug("TestAll End")
}
