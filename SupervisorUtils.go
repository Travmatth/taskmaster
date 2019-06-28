package main

import "os"

func (s *Supervisor) SafeAddProc(p *Proc) {
	s.mutex.Lock()
	s.procs[p.ID] = p
	s.mutex.Unlock()
}

func (s *Supervisor) SafeAddPid(p *Proc, process *os.Process) {
	s.mutex.Lock()
	s.processes[p.pid] = process
	s.mutex.Unlock()
}

func (s *Supervisor) SafeDeleteProc(p *Proc) {
	s.mutex.Lock()
	delete(s.procs, p.ID)
	s.mutex.Unlock()
}

func (s *Supervisor) SafeDeletePid(p *Proc) {
	s.mutex.Lock()
	delete(s.processes, p.pid)
	s.mutex.Unlock()
}

/*
DiffProcs identifies which processes are new and should be started,
which processes are changed and should be stopped and restarted,
and which processes are unchanged
*/
func (s *Supervisor) DiffProcs(configProcs []Proc) ([]Proc, []Proc) {
	newProcs := []Proc{}
	changedProcs := []Proc{}

	for _, proc := range configProcs {
		_, ok := s.procs[proc.ID]
		if ok == false {
			newProcs = append(newProcs, proc)
		} else {
			changedProcs = append(changedProcs, proc)
		}
	}
	return newProcs, changedProcs
}
