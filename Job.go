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
)

const (
	//RESTARTALWAYS signifies restart whenever process is stopped
	RESTARTALWAYS = iota
	//RESTARTNEVER signifies never restart process
	RESTARTNEVER
	//RESTARTUNEXPECTED signifies restart on unexpected exit
	RESTARTUNEXPECTED
)

type Job struct {
	ID            int
	args          []string
	AtLaunch      bool
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

func (j *Job) Start(wait bool) {
	j.mutex.Lock()
	if j.starting {
		j.mutex.Unlock()
		return
	}
	j.starting = true
	j.Stopped = false
	j.mutex.Unlock()
	var cond *sync.Cond
	done := false
	if wait {
		cond = sync.NewCond(&sync.Mutex{})
		cond.L.Lock()
	}
	go func() {
		for {
			if wait {
				cond.L.Lock()
			}
			j.Run(func() {
				done = true
				if wait {
					cond.L.Unlock()
					cond.Signal()
				}
			})
			if j.Stopped {
				Log.Info("Job", j.ID, "stopped by user, not restarting")
				break
			} else if j.process == nil {
				break
			} else if j.restartPolicy == RESTARTNEVER {
				Log.Info(j, "restart policy specifies do not restart")
				break
			} else if j.restartPolicy == RESTARTUNEXPECTED {
				end := true
				j.mutex.RLock()
				exit := j.state.ExitCode()
				if exit != j.ExpectedExit {
					Log.Info(j, "Encountered unexpected exit code", exit, ", restarting")
					end = false
				}
				j.mutex.RUnlock()
				if end {
					break
				}
			}
		}
		j.mutex.Lock()
		j.starting = false
		j.mutex.Unlock()
	}()
	if wait && !done {
		cond.Wait()
		cond.L.Unlock()
	}
}

//Run execs job
func (j *Job) Run(callback func()) {
	defer j.mutex.Unlock()
	j.mutex.Lock()
	if j.PIDExists() {
		Log.Info("Job", j.ID, "is already running")
		callback()
		return
	}
	j.StartTime = time.Now()
	atomic.StoreInt32(j.Restarts, 0)
	var once sync.Once
	callbackWrapper := func() {
		once.Do(callback)
	}
	for !j.Stopped {
		end := time.Now().Add(time.Duration(j.StartCheckup) * time.Second)
		j.ChangeStatus(PROCSTART)
		atomic.AddInt32(j.Restarts, 1)
		if err := j.CreateJob(); err != nil {
			if atomic.LoadInt32(j.Restarts) > j.MaxRestarts {
				errStr := fmt.Sprintf("Job %d failed to start with error: %s", j.ID, err)
				j.JobCreateFailure(callback, errStr)
				break
			} else {
				Log.Info("Job", j.ID, "failed to start with error:", err)
				j.ChangeStatus(PROCBACKOFF)
				continue
			}
		}
		monitorExited := int32(0)
		programExited := int32(0)
		if j.StartCheckup <= 0 {
			Log.Info(j, "Successfully Started")
			j.ChangeStatus(PROCRUNNING)
			go callbackWrapper()
		} else {
			go func() {
				j.MonitorProgramRunning(callbackWrapper, end, &monitorExited, &programExited)
			}()
		}
		j.mutex.Unlock()
		j.WaitForExit()
		atomic.StoreInt32(&programExited, 1)
		for j.StartCheckup > 0 && atomic.LoadInt32(&monitorExited) == 0 {
			time.Sleep(time.Duration(10) * time.Millisecond)
		}
		j.mutex.Lock()
		if j.Status == PROCRUNNING {
			j.ChangeStatus(PROCEXITED)
			break
		} else {
			j.ChangeStatus(PROCBACKOFF)
		}
		if atomic.LoadInt32(j.Restarts) >= j.MaxRestarts {
			j.JobCreateFailure(callback, fmt.Sprintf("Failed to start Job %d maximum retries reached", j.ID))
			break
		} else {
			Log.Info(j, "Start failed, restarting")
		}
	}
}

//CreateJob launches process
func (j *Job) CreateJob() error {
	defaultUmask := syscall.Umask(j.Umask)
	process, err := os.StartProcess(j.args[0], j.args, &os.ProcAttr{
		Dir:   j.WorkingDir,
		Env:   j.EnvVars,
		Files: j.Redirections,
	})
	if err != nil {
		return err
	}
	syscall.Umask(defaultUmask)
	j.process = process
	return nil
}

//WaitForExit waits for os.Process exit and saves returned ProcessState
func (j *Job) WaitForExit() {
	state, err := j.process.Wait()
	if err != nil {
		Log.Info("Job", j.ID, "error waiting for exit: ", err)
	} else if state != nil {
		Log.Info("Job", j.ID, "exited with status:", state)
	}
	j.mutex.Lock()
	j.state = state
	j.StopTime = time.Now()
	j.mutex.Unlock()
	select {
	case j.finishedCh <- struct{}{}:
	default:
	}
}

//PIDExists check the existence of given process
func (j *Job) PIDExists() bool {
	if j.process != nil && j.state != nil {
		//`man 2 kill`
		//If sig is 0, then no signal is sent, but error checking is still performed;
		//this can be used to check for the existence of a process ID or
		//process group ID.
		return j.process.Signal(Signals["SIGEXISTS"]) == nil
	}
	return false
}

//JobCreateFailure registers the failure of attempt to create job
func (j *Job) JobCreateFailure(callback func(), errStr string) {
	Log.Info("Creation of Job", j.ID, "failed:", errStr)
	callback()
}

// MonitorProgramRunning checks the run status of the program after exit
func (j *Job) MonitorProgramRunning(callback func(), end time.Time, monitor *int32, program *int32) {
	for time.Now().Before(end) && atomic.LoadInt32(program) == 0 {
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	atomic.StoreInt32(monitor, 1)
	defer j.mutex.Unlock()
	j.mutex.Lock()
	progState := atomic.LoadInt32(program)
	if progState == 0 && j.Status == PROCSTART {
		Log.Info("Job", j.ID, "Successfully Started")
		j.Status = PROCRUNNING
		callback()
	} else {
		Log.Debug(j, "monitor failed, program exit: ", progState, " with job status", j.Status)
	}
}

//ChangeStatus sets state
func (j *Job) ChangeStatus(state int) {
	j.Status = state
}

// Need to expand signal sending to -pid, target all in process group
func (j *Job) Stop(wait bool) {
	j.mutex.Lock()
	j.Stopped = true
	j.mutex.Unlock()
	go func() {
		j.mutex.RLock()
		if j.process != nil {
			Log.Info("Sending Signal", j.StopSignal, "to Job", j.ID)
			j.process.Signal(j.StopSignal)
		}
		j.mutex.RUnlock()
		select {
		case <-time.After(time.Duration(j.StopTimeout) * time.Second):
			Log.Info("Job", j.ID, "did not stop after timeout of ", j.StopTimeout, "seconds SIGKILL issued")
			if j.process != nil {
				j.process.Signal(Signals["SIGKILL"])
				<-j.finishedCh
			}
			break
		case <-j.finishedCh:
			break
		}
	}()
	if wait {
		for {
			j.mutex.RLock()
			if j.Status != PROCSTART && j.Status != PROCRUNNING && j.Status != PROCSTOPPING {
				j.mutex.RUnlock()
				break
			}
			j.mutex.RUnlock()
			time.Sleep(1 * time.Second)
		}
	}
}

func (j *Job) String() string {
	return fmt.Sprintf("Job %d", j.ID)
}
