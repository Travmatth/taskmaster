package main

import (
	"fmt"
	"os"
	"time"
)

// PerformStartCheckup called after process started, moves status: PROCSTART -> PROCRUNNING
func (p *Proc) PerformStartCheckup(s *Supervisor, process *os.Process) {
	timeout := time.Duration(p.StartCheckup) * time.Second
	time.Sleep(timeout)
	p.status = PROCRUNNING
	p.start = time.Now()
	fmt.Println("PerformStartCheckup ", p.ID, " is running ", p.status)
	go p.WaitForExit(s, process)
}

// WaitForExit catches process exit, removes process processes map, moves status: PROCRUNNING -> PROCSTOPPED
func (p *Proc) WaitForExit(s *Supervisor, process *os.Process) {
	_, _ = process.Wait()
	p.status = PROCSTOPPED
	delete(s.processes, p.pid)
	fmt.Println("WaitForExit ", p.ID, " has stopped ", p.status)
}
