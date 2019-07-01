package main

import (
	"os"
	"syscall"
)

func (s *Supervisor) WaitForExit() {
	s.wg.Wait()
	s.finishedCh <- struct{}{}
}

/*
DiffProcs identifies which processes are new and should be started,
which processes are changed and should be stopped and restarted,
and which processes are unchanged
*/
func (s *Supervisor) DiffJobs(configJobs []Job) ([]Job, []Job) {
	newJobs := []Job{}
	changedJobs := []Job{}

	for _, job := range configJobs {
		_, ok := s.Jobs[job.ID]
		if ok == false {
			newJobs = append(newJobs, job)
		} else {
			changedJobs = append(changedJobs, job)
		}
	}
	return newJobs, changedJobs
}

func (s *Supervisor) LaunchNewJob(job *Job) {
	defaultUmask := syscall.Umask(job.Umask)
	process, err := os.StartProcess(job.Args[0], job.Args, &os.ProcAttr{
		Dir:   job.WorkingDir,
		Env:   job.EnvVars,
		Files: job.Redirections,
	})
	if err != nil {
		Log.Info("Job ID", job.ID, " Failed to start with error: ", err)
		return
	}
	job.Status = PROCSTART
	syscall.Umask(defaultUmask)
	job.process = process
	job.Stopped = false
	job.mutex.Unlock()
	job.state, err = process.Wait()
	job.condition.Broadcast()
	job.mutex.Lock()
	if err != nil && job.Stopped == false {
		Log.Info("Job ID", job.ID, " Stopped with error: ", err)
	} else {
		Log.Info("Job ID", job.ID, " stopped")
	}
	job.process = nil
	s.wg.Done()
	job.mutex.Unlock()
}

func (s *Supervisor) KillJob(job *Job) error {
	return syscall.Kill(-job.process.Pid, Signals["SIGKILL"])
}

func (s *Supervisor) StopProcessGroup(job *Job) error {
	if job.process == nil {
		return nil
	}
	pgid, err := syscall.Getpgid(job.process.Pid)
	if err != nil {
		return err
	}
	pid := job.process.Pid
	if pgid == job.process.Pid {
		pid *= -1
	}
	if proc, err := os.FindProcess(pid); err != nil {
		return err
	} else {
		return proc.Signal(job.StopSignal)
	}
}
