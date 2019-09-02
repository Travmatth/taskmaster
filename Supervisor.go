package main

import (
	"os"
	"sync"
)

type Supervisor struct {
	Config  string
	LogFile string
	Mgr     *Manager
	lock    sync.Mutex
	restart bool
	SigCh   chan os.Signal
}

//NewSupervisor returns a Supervisor new struct
func NewSupervisor(file string, mgr *Manager, sigCh chan os.Signal, LogFile string) *Supervisor {
	return &Supervisor{
		Config:  file,
		Mgr:     mgr,
		SigCh:   sigCh,
		LogFile: LogFile,
	}
}

//DiffJobs sorts the given jobs into current, old, changed, and new slices
func (s *Supervisor) DiffJobs(jobs []*Job) ([]*Job, []*Job, []*Job, []*Job) {
	s.lock.Lock()
	defer s.lock.Unlock()
	currentJobs, oldJobs, changedJobs, newJobs := []*Job{}, []*Job{}, []*Job{}, []*Job{}
	for _, reloaded := range jobs {
		if job, ok := s.Mgr.Jobs[reloaded.ID]; !ok {
			Log.Info("Supervisor diffing next jobs: new", reloaded)
			newJobs = append(newJobs, reloaded)
		} else if diff := reloaded.cfg.Same(job.cfg); !diff {
			Log.Info("Supervisor diffing next jobs: changed", reloaded)
			changedJobs = append(changedJobs, job)
			newJobs = append(newJobs, reloaded)
			s.Mgr.RemoveJob(job.ID)
		} else {
			Log.Info("Supervisor diffing next jobs: current", job)
			currentJobs = append(currentJobs, job)
			s.Mgr.RemoveJob(job.ID)
		}
	}
	for _, job := range s.Mgr.Jobs {
		oldJobs = append(oldJobs, job)
		s.Mgr.RemoveJob(job.ID)
	}
	return currentJobs, oldJobs, changedJobs, newJobs
}

//Reload accepts a list of new jobs and diffs against current jobs
// to determine which to stop, start, remove, or continue unchanged
func (s *Supervisor) Reload(jobs []*Job, wait bool) error {
	curr, old, changed, next := s.DiffJobs(jobs)
	for _, job := range append(old, changed...) {
		job.Stop(wait)
	}
	s.AddMultiJobs(append(curr, next...))
	for _, job := range next {
		if job.AtLaunch {
			job.Start(wait)
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
	s.StopAllJobs(true)
}

//StartJob retrieves & starts a given job
func (s *Supervisor) StartJob(id int, wait bool) error {
	var job *Job
	var err error
	if job, err = s.Mgr.GetJob(id); err != nil {
		return err
	}
	job.Start(wait)
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
func (s *Supervisor) StartAllJobs(wait bool) {
	s.Mgr.lock.Lock()
	defer s.Mgr.lock.Unlock()
	ch := make(chan *Job)
	n := len(s.Mgr.Jobs)
	for _, job := range s.Mgr.Jobs {
		if job.AtLaunch {
			go func(job *Job) {
				job.Start(wait)
				ch <- job
			}(job)
		}
	}
	for i := 0; i < n; i++ {
		<-ch
	}
}

//StopAllJobs stops all jobs & waits for stop
func (s *Supervisor) StopAllJobs(wait bool) {
	s.Mgr.lock.Lock()
	defer s.Mgr.lock.Unlock()
	ch := make(chan *Job)
	n := len(s.Mgr.Jobs)
	for _, job := range s.Mgr.Jobs {
		go func(job *Job) {
			job.Stop(wait)
			ch <- job
		}(job)
	}
	for i := 0; i < n; i++ {
		<-ch
	}
}

//HasJob returns number of jobs being managed
func (s *Supervisor) HasJob(id int) bool {
	s.Mgr.lock.Lock()
	defer s.Mgr.lock.Unlock()
	_, ok := s.Mgr.Jobs[id]
	return ok
}

//ForAllJobs performs a callback on the managed jobs
func (s *Supervisor) ForAllJobs(f func(job *Job)) {
	s.Mgr.lock.Lock()
	defer s.Mgr.lock.Unlock()
	for _, job := range s.Mgr.Jobs {
		f(job)
	}
}
