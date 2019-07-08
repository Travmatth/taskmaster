package main

import (
	"reflect"
	"sync"
)

type Supervisor struct {
	Config  string
	Jobs    map[int]*Job
	Mgr     *Manager
	lock    sync.Mutex
	restart bool
}

func NewSupervisor(file string, mgr *Manager) *Supervisor {
	return &Supervisor{
		Config: file,
		Mgr:    mgr,
	}
}

func (s *Supervisor) DiffJobs(jobs []*Job) ([]*Job, []*Job, []*Job, []*Job) {
	defer s.lock.Unlock()
	s.lock.Lock()
	currentJobs, oldJobs, changedJobs, newJobs := []*Job{}, []*Job{}, []*Job{}, []*Job{}
	for _, cfg := range jobs {
		if job, ok := s.Jobs[cfg.ID]; !ok {
			newJobs = append(newJobs, cfg)
		} else if diff := reflect.DeepEqual(cfg.cfg, job.cfg); diff {
			changedJobs = append(changedJobs, cfg)
		} else {
			currentJobs = append(currentJobs, cfg)
		}
		s.Mgr.RemoveJob(cfg)
	}
	for _, job := range s.Jobs {
		oldJobs = append(oldJobs, job)
		s.Mgr.RemoveJob(job)
	}
	return currentJobs, oldJobs, changedJobs, newJobs
}

func (s *Supervisor) Reload(jobs []*Job) error {
	defer s.Mgr.lock.Unlock()
	s.Mgr.lock.Lock()
	curr, old, changed, next := s.DiffJobs(jobs)
	for _, job := range curr {
		s.Mgr.AddJob(job)
	}
	for _, job := range old {
		job.Stop(false)
	}
	for _, job := range changed {
		s.Mgr.RestartJob(job)
	}
	for _, job := range next {
		s.Mgr.AddJob(job)
		if job.AtLaunch {
			s.Mgr.StartJob(job)
		}
	}
	return nil
}

func (s *Supervisor) WaitForExit() {
	s.Mgr.StopAllJobs()
}

func (s *Supervisor) StartAllJobs() {
	ch := make(chan *Job)
	s.Mgr.lock.Lock()
	n := len(s.Mgr.Jobs)
	for _, job := range s.Mgr.Jobs {
		go func(job *Job) {
			job.Start(true)
			ch <- job
		}(job)
	}
	for i := 0; i < n; i++ {
		<-ch
	}
	s.Mgr.lock.Unlock()
}

func (s *Supervisor) StopAllJobs() {
	ch := make(chan *Job)
	s.Mgr.lock.Lock()
	n := len(s.Mgr.Jobs)
	for _, job := range s.Mgr.Jobs {
		go func(job *Job) {
			job.Stop(true)
			ch <- job
		}(job)
	}
	for i := 0; i < n; i++ {
		<-ch
	}
	s.Mgr.lock.Unlock()
}
