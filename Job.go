package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

const (
	// PROCSTOPPED signifies process not currently running
	PROCSTOPPED = iota
	// PROCRUNNING signifies process currently running
	PROCRUNNING
	// PROCSTART signifies process beginning its start sequence
	PROCSTART
	// PROCEXITED signifies process has ended
	PROCEXITED
	PROCBACKOFF
	PROCSTOPPING
)

const (
	// RESTARTALWAYS signifies restart whenever process is stopped
	RESTARTALWAYS = iota
	// RESTARTNEVER signifies never restart process
	RESTARTNEVER
	// RESTARTUNEXPECTED signifies restart on unexpected exit
	RESTARTUNEXPECTED
)

type Job struct {
	ID            int
	args          []string
	AtLaunch      bool
	restartPolicy int
	ExpectedExit  os.Signal
	StartCheckup  int
	MaxRestarts   *int32
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
	Log           *logging.Logger
	starting      bool
}

func (j *Job) Start(wait bool) {
	fmt.Println("Attempting to start job", j.ID)
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
	// ch := make(chan struct{})
	go func() {
		fmt.Println("Job", j.ID, "Run callback()")
		for {
			if wait {
				cond.L.Lock()
			}
			j.Run(func() {
				// ch <- struct{}{}
				done = true
				if wait {
					cond.L.Unlock()
					cond.Signal()
				}
			})
			if j.Stopped {
				j.Log.Debug("Job", j.ID, "stopped by user, not restarting")
				break
			} else if j.restartPolicy != RESTARTALWAYS {
				j.Log.Debug("Job", j.ID, "restart policy specifies do not restart")
				break
			}
		}
		j.mutex.Lock()
		j.starting = false
		j.mutex.Unlock()
	}()
	// <-ch
	fmt.Println("Job", j.ID, "post callback()")
	if wait && !done {
		cond.Wait()
		cond.L.Unlock()
	}
}

func (j *Job) Run(callback func()) {
	defer j.mutex.Unlock()
	j.mutex.Lock()
	fmt.Println("Run called on Job", j.ID)
	if j.PIDExists() {
		j.Log.Info("Job", j.ID, "is already running")
		callback()
		return
	}
	j.StartTime = time.Now()
	var once sync.Once
	callbackWrapper := func() {
		once.Do(callback)
	}
	for !j.Stopped {
		end := time.Now().Add(time.Duration(j.StartCheckup) * time.Second)
		j.ChangeStatus(PROCSTART)
		atomic.AddInt32(j.MaxRestarts, -1)
		err := j.CreateJob()
		if err != nil {
			if atomic.LoadInt32(j.MaxRestarts) <= 0 {
				errStr := fmt.Sprintf("Job %d failed to start after retries with error: %s", j.ID, err)
				j.JobCreateFailure(callback, errStr)
				break
			} else {
				j.Log.Info("Job", j.ID, "failed to start with error:", err)
				continue
			}
		}
		monitorExited := int32(0)
		programExited := int32(0)
		if j.StartCheckup <= 0 {
			j.Log.Info("Job", j.ID, "Successfully Started")
			j.ChangeStatus(PROCRUNNING)
			go callbackWrapper()
		} else {
			go func() {
				j.MonitorProgramRunning(end, &monitorExited, &programExited)
				callbackWrapper()
			}()
		}
		fmt.Println("Job", j.ID, "waiting for program exit")
		j.mutex.Unlock()
		j.WaitForExit()
		j.mutex.Lock()
		if j.Status == PROCRUNNING {
			j.ChangeStatus(PROCEXITED)
			j.Log.Info("Job", j.ID, "stopped")
			break
		} else {
			j.ChangeStatus(PROCBACKOFF)
		}
		if atomic.LoadInt32(j.MaxRestarts) <= 0 {
			j.JobCreateFailure(callback, fmt.Sprintf("Failed to start Job %d maximum retries reached", j.ID))
			break
		}
	}
}

//WaitForExit waits for os.Process exit and saves returned ProcessState
func (j *Job) WaitForExit() {
	fmt.Println("Waiting for job", j.ID, "exit")
	state, err := j.process.Wait()
	fmt.Println("Waiting for job", j.ID, "exit completed")
	if err != nil {
		j.Log.Info("Job", j.ID, "error waiting for exit: ", err)
	} else if state != nil {
		j.Log.Info("Job", j.ID, "stopped with status:", state)
	} else {
		j.Log.Info("Job", j.ID, "stopped")
	}
	defer j.mutex.Unlock()
	j.mutex.Lock()
	j.state = state
	j.StopTime = time.Now()
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

//JobCreateFailure registers the failure of attempt to create job
func (j *Job) JobCreateFailure(callback func(), errStr string) {
	j.Log.Info("Creation of Job", j.ID, "failed:", errStr)
	callback()
}

// MonitorProgramRunning checks the run status of the program after exit
func (j *Job) MonitorProgramRunning(end time.Time, monitor *int32, program *int32) {
	for time.Now().Before(end) && atomic.LoadInt32(program) == 0 {
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	atomic.StoreInt32(monitor, 1)
	defer j.mutex.Unlock()
	j.mutex.Lock()
	if atomic.LoadInt32(program) == 0 && j.Status == PROCSTART {
		fmt.Println("Job", j.ID, "started")
		j.Log.Info("Job", j.ID, "started")
	}
}

func (j *Job) ChangeStatus(state int) {
	j.Status = state
}

// Need to expand signal sending to -pid, target all in process group
func (j *Job) Stop(wait bool) {
	j.mutex.Lock()
	fmt.Println("Setting job", j.ID, "to stopped")
	j.Stopped = true
	j.mutex.Unlock()
	go func() {
		stopped := false
		j.mutex.Lock()
		if j.process != nil {
			fmt.Println("Sending stop sig to job", j.ID)
			j.process.Signal(j.StopSignal)
		}
		j.mutex.Unlock()
		for time.Now().Add(time.Duration(j.StopTimeout) * time.Second).After(time.Now()) {
			if j.Status != PROCSTART && j.Status != PROCRUNNING && j.Status != PROCSTOPPING {
				j.Log.Info("Job", j.ID, "Stopped Successfully")
				stopped = true
				break
			}
			time.Sleep(1 * time.Second)
		}
		if !stopped {
			j.Log.Info("Job", j.ID, "did not stop, SIGKILL issued")
			if j.process != nil {
				j.process.Signal(Signals["SIGKILL"])
			}
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
