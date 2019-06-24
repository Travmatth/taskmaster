package main

import (
	"os"
	"time"
)

// PerformStartCheckup called after process started, moves status: PROCSTART -> PROCRUNNING
func (p *Proc) PerformStartCheckup(events chan ProcEvent, process *os.Process) {
	Log.Debug("PerformStartCheckup start")
	time.Sleep(time.Duration(p.StartCheckup) * time.Second)
	events <- ProcEvent{SETPROCRUNNING, p}
	Log.Debug("PerformStartCheckup end")
	go p.WaitForExit(events, process)
}

// WaitForExit catches process exit, removes process processes map, moves status: PROCRUNNING -> PROCSTOPPED
func (p *Proc) WaitForExit(events chan ProcEvent, process *os.Process) {
	Log.Debug("WaitForExit start")
	_, _ = process.Wait()
	events <- ProcEvent{SETPROCSTOPPED, p}
	Log.Debug("WaitForExit end")
}
