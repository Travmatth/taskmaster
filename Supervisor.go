package main

import (
	"sync"
	"time"
)

// Supervisor struct contains array of Proc structs
type Supervisor struct {
	Jobs        map[int]*Job
	PIDS        map[int]int
	wg          sync.WaitGroup
	errCh       chan error
	finishedCh  chan struct{}
	stopChan    chan Job
	startChan   chan Job
	restartChan chan Job
}

func NewSupervisor() *Supervisor {
	var s Supervisor
	s.Jobs = make(map[int]*Job)
	s.PIDS = make(map[int]int)
	s.finishedCh = make(chan struct{})
	s.stopCh = make(chan Job)
	s.startCh = make(chan Job)
	s.restartCh = make(chan Job)
	return &s
}

func (s *Supervisor) StartJob(job *Job) error {
	job.mutex.Lock()
	if job.process != nil {
		job.mutex.Unlock()
		Log.Debug("Job ", job.ID, " already started")
		return nil
	}
	s.wg.Add(1)
	go s.LaunchNewJob(job)
	return nil
}

func (s *Supervisor) StopJob(job *Job) error {
	defer job.mutex.Unlock()
	job.mutex.Lock()
	if job.process == nil {
		return nil
	}
	job.Stopped = true
	if err := s.StopProcessGroup(job); err != nil {
		return err
	} else {
		timer := time.AfterFunc(time.Duration(job.StopTimeout)*time.Second, func() {
			defer job.mutex.Unlock()
			job.mutex.Lock()
			if job.cmd != nil {
				err = s.KillJob(job)
			}
		})
		job.condition.Wait()
		timer.Stop()
		return err
	}
}

func (s *Supervisor) StopAllJobs() error {
	for job := range s.Jobs {
		s.StopJob(job)
	}
	return nil
}

func (s *Supervisor) StartAllJobs([]*Job) error {
	for job := range jobs {
		s.StartJob(job)
	}
	go s.WaitForExit()
	for {
		select {
		case job := <-s.stopChan:
			if err := s.StopJob(job); err != nil {
				Log.Info("Error stopping job ", job.ID, ": ", err)
			}
		case <-s.finishedCh:
			return s.StopAllJobs()
		}
	}
	return nil
}
