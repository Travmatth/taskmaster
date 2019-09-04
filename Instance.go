package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	// PROCSTOPPED signifies process not currently running
	PROCSTOPPED = iota
	// PROCRUNNING signifies process currently running
	PROCRUNNING
	//PROCSTART signifies process beginning its start sequence
	PROCSTART
	//PROCEXITED signifies process has ended
	PROCEXITED
	//PROCBACKOFF signifies that process in intermediate state, should not alter
	PROCBACKOFF
	//PROCSTOPPING signifies in process of stopping
	PROCSTOPPING
	//PROCSTARTFAIL signifies process could not start successfully
	PROCSTARTFAIL
)

const (
	//RESTARTALWAYS signifies restart whenever process is stopped
	RESTARTALWAYS = iota
	//RESTARTNEVER signifies never restart process
	RESTARTNEVER
	//RESTARTUNEXPECTED signifies restart on unexpected exit
	RESTARTUNEXPECTED
)

type Instance struct {
	JobID         int
	InstanceID    int
	args          []string
	restartPolicy int
	ExpectedExit  int
	StartCheckup  int
	Restarts      *int32
	MaxRestarts   int32
	StopSignal    os.Signal
	StopTimeout   int
	EnvVars       []string
	WorkingDir    string
	Umask         int
	StartTime     time.Time
	StopTime      time.Time
	Status        int
	Redirections  []*os.File
	Stopped       bool
	process       *os.Process
	mutex         sync.RWMutex
	condition     *sync.Cond
	state         *os.ProcessState
	cfg           *JobConfig
	starting      bool
	finishedCh    chan struct{}
}

/*
StartInstance manages the execution of a process by launching a process by calling instance.Run()
and rerunning the process after exit if restartPolicy == always
or restartPolicy == unexpected && the exit code does not match the exit code specified in the config
*/
func (i *Instance) StartInstance(wait bool) {
	i.mutex.Lock()
	if i.starting {
		i.mutex.Unlock()
		return
	}
	i.starting = true
	i.Stopped = false
	i.mutex.Unlock()
	var cond *sync.Cond
	done := false
	if wait {
		cond = sync.NewCond(&sync.Mutex{})
		cond.L.Lock()
	}
	go func() {
	Rerun:
		for {
			if wait {
				cond.L.Lock()
			}
			i.Run(func() {
				done = true
				if wait {
					cond.L.Unlock()
					cond.Signal()
				}
			})
			for i.Status == PROCBACKOFF {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
			if !i.shouldRerunInstance() {
				break Rerun
			}
		}
		i.mutex.Lock()
		i.starting = false
		i.mutex.Unlock()
	}()
	if wait && !done {
		cond.Wait()
		cond.L.Unlock()
	}
}

/*
shouldRerunInstance determines whether the process has exited due to a failed start
or a successful execution, and if the process should be rerun (when restartPolicy == always or unexpected)
*/
func (i *Instance) shouldRerunInstance() bool {
	switch {
	case i.Stopped:
		Log.Info(i, ": stopped by user, not restarting")
		return false
	case i.process == nil || i.Status == PROCSTARTFAIL:
		return false
	case i.restartPolicy == RESTARTNEVER:
		Log.Info(i, ": restart policy specifies do not restart")
		return false
	case i.restartPolicy == RESTARTUNEXPECTED:
		end := true
		i.mutex.RLock()
		exit := i.state.ExitCode()
		if exit != i.ExpectedExit {
			Log.Info(i, ": Encountered unexpected exit code", exit, ", restarting")
			end = false
		}
		i.mutex.RUnlock()
		if end {
			return false
		}
	}
	return true
}

/*
Run launches the process and monitors the start, restarting if start has failed
on successful start, waits for process to complete and sends a finished message on finishedCh
*/
func (i *Instance) Run(callback func()) {
	defer i.mutex.Unlock()
	i.mutex.Lock()
	if i.PIDExists() {
		Log.Info(i, ": already running")
		callback()
		return
	}
	i.StartTime = time.Now()
	atomic.StoreInt32(i.Restarts, 0)
	var once sync.Once
	for !i.Stopped {
		i.ChangeStatus(PROCSTART)
		atomic.AddInt32(i.Restarts, 1)
		if err := i.CreateJob(); err != nil {
			restarts := atomic.LoadInt32(i.Restarts)
			if restarts > i.MaxRestarts {
				errStr := fmt.Sprintf("failed to start with error: %s", err)
				Log.Info(i, ": Creation failed:", errStr)
				callback()
				break
			} else {
				Log.Info(i, ": failed to start with error:", err)
				i.ChangeStatus(PROCBACKOFF)
				continue
			}
		}
		i.manageRunningProgram(func() {
			once.Do(callback)
		})
		if !i.shouldRestartInstance(callback) {
			break
		}
	}
}

/*
manageRunningProgram watches the running process, notifying the parent once it has successfully started
and then waiting for process exit
*/
func (i *Instance) manageRunningProgram(callbackWrapper func()) {
	end := time.Now().Add(time.Duration(i.StartCheckup) * time.Second)
	monitorExited := int32(0)
	programExited := int32(0)
	if i.StartCheckup <= 0 {
		Log.Info(i, ": Successfully Started with no start checkup")
		i.ChangeStatus(PROCRUNNING)
		go callbackWrapper()
	} else {
		go func() {
			i.startCheckup(callbackWrapper, end, &monitorExited, &programExited)
		}()
	}
	i.mutex.Unlock()
	i.WaitForExit()
	atomic.StoreInt32(&programExited, 1)
	for i.StartCheckup > 0 && atomic.LoadInt32(&monitorExited) == 0 {
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
}

//shouldRestartInstance determines whether process has successfully started, or should restart
func (i *Instance) shouldRestartInstance(callback func()) bool {
	i.mutex.Lock()
	if i.Status == PROCRUNNING {
		i.ChangeStatus(PROCEXITED)
		select {
		case i.finishedCh <- struct{}{}:
		default:
		}
		return false
	}
	i.ChangeStatus(PROCBACKOFF)
	if atomic.LoadInt32(i.Restarts) >= i.MaxRestarts {
		i.ChangeStatus(PROCSTARTFAIL)
		Log.Info(i, ": Creation failed:", fmt.Sprintf("Failed to start maximum retries reached"))
		callback()
		return false
	}
	Log.Info(i, ": Start failed, restarting")
	return true
}

//CreateJob creates & launches process
func (i *Instance) CreateJob() error {
	defaultUmask := syscall.Umask(i.Umask)
	defer syscall.Umask(defaultUmask)
	process, err := os.StartProcess(i.args[0], i.args, &os.ProcAttr{
		Dir:   i.WorkingDir,
		Env:   i.EnvVars,
		Files: i.Redirections,
	})
	if err != nil {
		return err
	}
	i.process = process
	return nil
}

//WaitForExit waits for os.Process exit and saves returned ProcessState
func (i *Instance) WaitForExit() {
	state, err := i.process.Wait()
	if err != nil {
		Log.Info(i, ": error waiting for exit: ", err)
	} else if state != nil {
		Log.Info(i, ": exited with status:", state)
	}
	i.mutex.Lock()
	i.state = state
	i.StopTime = time.Now()
	i.mutex.Unlock()
}

//PIDExists check the existence of given process
func (i *Instance) PIDExists() bool {
	if i.process != nil && i.state != nil {
		//`man 2 kill`
		//If sig is 0, then no signal is sent, but error checking is still performed;
		//this can be used to check for the existence of a process ID or
		//process group ID.
		return i.process.Signal(Signals["SIGEXISTS"]) == nil
	}
	return false
}

//startCheckup checks that the process has successfully started after the specified start checkup time
func (i *Instance) startCheckup(callback func(), end time.Time, monitor *int32, program *int32) {
	for time.Now().Before(end) && atomic.LoadInt32(program) == 0 {
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	atomic.StoreInt32(monitor, 1)
	defer i.mutex.Unlock()
	i.mutex.Lock()
	progState := atomic.LoadInt32(program)
	if progState == 0 && i.Status == PROCSTART {
		Log.Info(i, ": Successfully Started after", i.StartCheckup, "second(s)")
		i.Status = PROCRUNNING
		callback()
	} else {
		Log.Info(i, ": monitor failed, program exit: ", progState, " with job status", i.Status)
	}
}

//ChangeStatus sets state
func (i *Instance) ChangeStatus(state int) {
	i.Status = state
}

/*
StopInstance is used to stop the process by sending the specified stop signal
or the SIGKILL signal if the stop signal is not received
*/
func (i *Instance) StopInstance(wait bool) {
	i.mutex.Lock()
	i.Stopped = true
	i.mutex.Unlock()
	go i.stopTimeout()
	if wait {
		for {
			i.mutex.RLock()
			if i.Status != PROCSTART && i.Status != PROCRUNNING && i.Status != PROCSTOPPING {
				i.mutex.RUnlock()
				break
			}
			i.mutex.RUnlock()
			time.Sleep(1 * time.Second)
		}
	}
}

/*
stopTimeout signals the process to stop using the specified signal
if wait() is not called then a SIGKILL is sent to the process
*/
func (i *Instance) stopTimeout() {
	i.mutex.RLock()
	if i.process != nil {
		Log.Info(i, ": Sending Signal", i.StopSignal)
		i.process.Signal(i.StopSignal)
	}
	i.mutex.RUnlock()
	select {
	case <-time.After(time.Duration(i.StopTimeout) * time.Second):
		Log.Info(i, ": did not stop after timeout of ", i.StopTimeout, "seconds SIGKILL issued")
		if i.process != nil {
			i.process.Signal(Signals["SIGKILL"])
			<-i.finishedCh
		}
	case <-i.finishedCh:
	}
}

//GetStatus return status of the process
func (i *Instance) GetStatus() string {
	status := ""
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	switch i.Status {
	case PROCSTOPPED:
		status = "stopped"
	case PROCRUNNING:
		status = "running"
	case PROCSTART:
		status = "start"
	case PROCEXITED:
		status = "exited"
	case PROCBACKOFF:
		status = "backoff"
	case PROCSTOPPING:
		status = "stopping"
	case PROCSTARTFAIL:
		status = "start failed"
	}
	return status
}

func (i *Instance) String() string {
	return fmt.Sprintf("Job %d Instance %d", i.JobID, i.InstanceID)
}
