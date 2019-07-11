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

//Reload accepts a list of new jobs and diffs against current jobs
// to determine which to stop, start, remove, or continue unchanged
func (s *Supervisor) Reload(jobs []*Job) error {
	curr, old, changed, next := s.DiffJobs(jobs)
	for _, job := range curr {
		s.Mgr.AddSingleJob(job)
	}
	for _, job := range old {
		job.Stop(false)
	}
	for _, job := range changed {
		s.Mgr.RestartJob(job)
	}
	for _, job := range next {
		s.Mgr.AddSingleJob(job)
		if job.AtLaunch {
			job.Start(false)
		}
	}
	return nil
}

//WaitForExit waits for exit
func (s *Supervisor) WaitForExit() {
	s.StopAllJobs()
}

//StartJob retrieves & starts a given job
func (s *Supervisor) StartJob(id int) error {
	var job *Job
	var err error
	if job, err = s.Mgr.GetJob(id); err != nil {
		return err
	}
	job.Start(true)
	return nil
}

//StopJob retrieves & stops a given job
func (s *Supervisor) StopJob(id int) error {
	var job *Job
	var err error
	if job, err = s.Mgr.GetJob(id); err != nil {
		return err
	}
	job.Stop(true)
	return nil
}

//StartAllJobs starts all jobs & waits for start
func (s *Supervisor) StartAllJobs() {
	ch := make(chan *Job)
	s.Mgr.lock.Lock()
	n := len(s.Mgr.Jobs)
	for _, job := range s.Mgr.Jobs {
		if job.AtLaunch {
			go func(job *Job) {
				job.Start(true)
				ch <- job
			}(job)
		}
	}
	for i := 0; i < n; i++ {
		<-ch
	}
	s.Mgr.lock.Unlock()
}

//StopAllJobs stops all jobs & waits for stop
func (s *Supervisor) StopAllJobs() {
	Log.Info("StopAllJobs start")
	ch := make(chan *Job)
	s.Mgr.lock.Lock()
	n := len(s.Mgr.Jobs)
	for _, job := range s.Mgr.Jobs {
		go func(job *Job) {
			job.Stop(true)
			Log.Info("Stop end")
			ch <- job
		}(job)
	}
	for i := 0; i < n; i++ {
		<-ch
	}
	s.Mgr.lock.Unlock()
	Log.Info("StopAllJobs end")
}
