package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/go-cmp/cmp"
)

// Supervisor struct contains array of Proc structs
type Supervisor struct {
	procs     map[int]Proc
	processes map[int]*os.Process
}

/*
DiffProcs identifies which processes are new and should be started,
which processes are changed and should be stopped and restarted,
and which processes are unchanged
*/
func (s *Supervisor) DiffProcs(configProcs []Proc) ([]Proc, []Proc, []Proc) {
	fmt.Println("DiffProcs Start")
	newProcs := []Proc{}
	sameProcs := []Proc{}
	changedProcs := []Proc{}

	for _, proc := range configProcs {
		val, ok := s.procs[proc.ID]
		if ok == false {
			fmt.Println("New proc detected")
			newProcs = append(newProcs, proc)
		} else if cmp.Equal(proc, s.procs[proc.ID]) {
			fmt.Println("Same proc detected")
			sameProcs = append(sameProcs, val)
		} else {
			fmt.Println("Changed proc detected")
			changedProcs = append(changedProcs, proc)
		}
	}
	fmt.Println("DiffProcs End")
	return newProcs, sameProcs, changedProcs
}

// StartAll iterates through given processes, launching each
func (s *Supervisor) StartAll(configProcs []Proc) {
	fmt.Println("StartAll Start")
	newProcs, sameProcs, changedProcs := s.DiffProcs(configProcs)
	nextProcs := make(map[int]Proc)
	for _, proc := range sameProcs {
		nextProcs[proc.ID] = proc
	}
	for _, proc := range newProcs {
		nextPid, err := proc.Start(s)
		if err == nil {
			nextProcs[nextPid] = proc
		} else {
			fmt.Println(err)
		}
	}
	for _, proc := range changedProcs {
		nextPid, err := proc.Restart(s)
		if err == nil {
			nextProcs[nextPid] = proc
		} else {
			fmt.Print(err)
		}
	}
	s.procs = nextProcs
	fmt.Println("StartAll End")
}

// TestAll launches proc test - starts all proceses, then kills all processes
func (s *Supervisor) TestAll(configProcs []Proc) {
	fmt.Println("TestAll Start")
	s.StartAll(configProcs)
	time.Sleep(time.Duration(5) * time.Second)
	for _, p := range s.procs {
		p.Kill(s)
	}
	fmt.Println("TestAll End")
}
