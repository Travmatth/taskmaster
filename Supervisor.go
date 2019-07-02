package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/op/go-logging"
)

// Supervisor struct contains array of Proc structs and methods to communicate with them
type Supervisor struct {
	Jobs       map[int]*Job
	PIDS       map[int]int
	wg         sync.WaitGroup
	errCh      chan error
	finishedCh chan struct{}
	stopCh     chan int
	Log        *logging.Logger
}

// NewSupervisor creates and returns new supervisor struct
func NewSupervisor(log *logging.Logger) *Supervisor {
	var s Supervisor
	s.Jobs = make(map[int]*Job)
	s.PIDS = make(map[int]int)
	s.finishedCh = make(chan struct{})
	s.stopCh = make(chan int)
	s.Log = log
	return &s
}

// StopJob stops given job
func (s *Supervisor) StopJob(id int) error {
	var err error
	job, ok := s.Jobs[id]
	if !ok {
		return fmt.Errorf("Attempting to stop unknown Job: %d", id)
	}
	defer job.mutex.Unlock()
	job.mutex.Lock()
	if job.process == nil {
		return nil
	}
	job.Stopped = true
	if err = s.StopProcessGroup(job); err != nil {
		s.Log.Error("here")
		return err
	}
	// https://github.com/golang/go/issues/9578
	timer := time.AfterFunc(time.Duration(job.StopTimeout)*time.Second, func() {
		defer job.mutex.Unlock()
		job.mutex.Lock()
		if job, ok := s.Jobs[job.ID]; ok && job.process != nil {
			s.Log.Error("there")
			err = s.KillJob(job)
		}
	})
	job.condition.Wait()
	timer.Stop()
	return err
}

// StopAllJobs attempts to stop all running jobs
func (s *Supervisor) StopAllJobs() error {
	for _, job := range s.Jobs {
		s.StopJob(job.ID)
	}
	return nil
}

// StartJob launches given job
func (s *Supervisor) StartJob(job *Job) error {
	job.mutex.Lock()
	if job.process != nil {
		job.mutex.Unlock()
		s.Log.Debug("Job ", job.ID, " already started")
		return nil
	}
	s.wg.Add(1)
	go s.LaunchNewJob(job)
	s.Log.Debug("Finished StartJob")
	return nil
}

// StartAllJobs starts given jobs and
func (s *Supervisor) StartAllJobs(jobs []*Job) error {
	s.Log.Debug("StartAllJobs, starting", len(jobs), "jobs(s)")
	for _, job := range jobs {
		s.Jobs[job.ID] = job
		s.StartJob(job)
	}
	go s.WaitForExit()
	for {
		select {
		case id := <-s.stopCh:
			s.Log.Debug("Recieved stop message: Job", id)
			if err := s.StopJob(id); err != nil {
				s.Log.Info("Error stopping job ", id, ": ", err)
			} else {
				s.Log.Info("Stopped job ", id)
			}
			s.Log.Info("Stopped job end")
		case <-s.finishedCh:
			return nil
		}
	}
}
