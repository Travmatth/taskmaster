package main

import (
	"fmt"
	"os"

	"github.com/google/go-cmp/cmp"
)

// Supervisor struct contains array of Proc structs
type Supervisor struct {
	procs map[int]Proc
	pids  map[int]*os.Process
}

/*
DiffProcs identifies which processes are new and should be started,
which processes are changed and should be stopped and restarted,
and which processes are unchanged
*/
func (s *Supervisor) DiffProcs(configProcs []Proc) ([]Proc, []Proc, []Proc) {
	newProcs := []Proc{}
	sameProcs := []Proc{}
	changedProcs := []Proc{}

	for _, proc := range configProcs {
		val, ok := s.procs[proc.ID]
		switch ok {
		case ok && cmp.Equal(proc, s.procs[proc.ID]):
			sameProcs = append(sameProcs, val)
		case ok:
			changedProcs = append(changedProcs, proc)
		default:
			newProcs = append(newProcs, proc)
		}
	}
	return newProcs, sameProcs, changedProcs
}

// StartAll iterates through given processes, launching each
func (s *Supervisor) StartAll(configProcs []Proc) {
	newProcs, sameProcs, changedProcs := s.DiffProcs(configProcs)
	nextProcs := make(map[int]Proc)
	for _, proc := range sameProcs {
		nextProcs[proc.ID] = proc
	}
	for _, proc := range newProcs {
		nextPid, err := proc.Start(s.pids)
		if err == nil {
			s.procs[nextPid] = proc
		} else {
			fmt.Print(err)
		}
	}
	for _, proc := range changedProcs {
		nextPid, err := proc.Restart(s.procs[proc.ID].pid, s.pids)
		if err == nil {
			s.procs[nextPid] = proc
		} else {
			fmt.Print(err)
		}
	}
	s.procs = nextProcs
}
