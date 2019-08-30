package main

import (
	"sync"
)

type Supervisor struct {
	Config  string
	Mgr     *Manager
	lock    sync.Mutex
	restart bool
}

//NewSupervisor returns a Supervisor new struct
func NewSupervisor(file string, mgr *Manager) *Supervisor {
	return &Supervisor{
		Config: file,
		Mgr:    mgr,
	}
}

//DiffJobs sorts the given jobs into current, old, changed, and new slices
func (s *Supervisor) DiffJobs(jobs []*Job) ([]*Job, []*Job, []*Job, []*Job) {
	defer s.lock.Unlock()
	s.lock.Lock()
	currentJobs, oldJobs, changedJobs, newJobs := []*Job{}, []*Job{}, []*Job{}, []*Job{}
	for _, cfg := range jobs {
		if job, ok := s.Mgr.Jobs[cfg.ID]; !ok {
			Log.Info("Supervisor diffing next jobs: new", cfg)
			newJobs = append(newJobs, cfg)
		} else if diff := cfg.cfg.Same(job.cfg); !diff {
			Log.Info("Supervisor diffing next jobs: changed", cfg)
			changedJobs = append(changedJobs, cfg)
		} else {
			Log.Info("Supervisor diffing next jobs: current", cfg)
			currentJobs = append(currentJobs, cfg)
			s.Mgr.RemoveJob(cfg.ID)
		}
		s.Mgr.RemoveJob(cfg.ID)
	}
	for _, job := range s.Mgr.Jobs {
		oldJobs = append(oldJobs, job)
		s.Mgr.RemoveJob(job.ID)
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
		s.Mgr.RestartJob(job.ID)
	}
	for _, job := range next {
		s.Mgr.AddSingleJob(job)
		if job.AtLaunch {
			job.Start(false)
		}
	}
	return nil
}

//AddMultiJobs add multiple jobs to manager
func (s *Supervisor) AddMultiJobs(jobs []*Job) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Mgr.AddMultiJobs(jobs)
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

//GetJob returns the job with the given id
func (s *Supervisor) GetJob(id int) (*Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.Mgr.GetJob(id)
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
