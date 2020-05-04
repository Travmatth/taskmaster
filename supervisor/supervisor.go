package supervisor

import (
	"os"
	"sync"

	. "github.com/Travmatth/taskmaster/job"
	. "github.com/Travmatth/taskmaster/log"
)

/*
 * Supervisor models the users manipulation of jobs
 */
type Supervisor struct {
	Config  string
	LogFile string
	Mgr     *Manager
	lock    sync.Mutex
	restart bool
	SigCh   chan os.Signal
}

/*
 * NewSupervisor returns a new Supervisor struct
 */
func NewSupervisor(cfg, logFile string,
	mgr *Manager, sigCh chan os.Signal) *Supervisor {
	return &Supervisor{
		Config:  cfg,
		Mgr:     mgr,
		SigCh:   sigCh,
		LogFile: logFile,
	}
}

/*
 * DiffJobs sorts the given jobs into current, old, changed, and new slices
 */
func (s *Supervisor) DiffJobs(jobs []*Job) ([]*Job, []*Job, []*Job, []*Job) {
	defer s.lock.Unlock()
	s.lock.Lock()
	current, old, changed, new := []*Job{}, []*Job{}, []*Job{}, []*Job{}
	for _, reloaded := range jobs {
		if job, ok := s.Mgr.Jobs[reloaded.ID]; !ok {
			Log.Info("Supervisor diffing next jobs: new", reloaded)
			new = append(new, reloaded)
		} else if diff := reloaded.Cfg.Same(job.Cfg); !diff {
			Log.Info("Supervisor diffing next jobs: changed", reloaded)
			changed = append(changed, job)
			new = append(new, reloaded)
			s.Mgr.RemoveJob(job.ID)
		} else {
			Log.Info("Supervisor diffing next jobs: current", job)
			current = append(current, job)
			s.Mgr.RemoveJob(job.ID)
		}
	}
	for _, job := range s.Mgr.Jobs {
		old = append(old, job)
		s.Mgr.RemoveJob(job.ID)
	}
	return current, old, changed, new
}

/*
 * Reload accepts a list of new jobs and diffs against current jobs
 * to determine which to stop, start, remove, or continue unchanged
 */
func (s *Supervisor) Reload(jobs []*Job, wait bool) error {
	current, old, changed, next := s.DiffJobs(jobs)
	for _, job := range append(old, changed...) {
		job.Stop(wait)
	}
	s.AddMultiJobs(append(current, next...))
	for _, job := range next {
		if job.AtLaunch {
			job.Start(wait)
		}
	}
	return nil
}

/*
 * AddMultiJobs add multiple jobs to manager
 */
func (s *Supervisor) AddMultiJobs(jobs []*Job) {
	defer s.lock.Unlock()
	s.lock.Lock()
	s.Mgr.AddMultiJobs(jobs)
}

/*
 * WaitForExit waits for exit
 */
func (s *Supervisor) WaitForExit() {
	s.StopAllJobs(true)
}

/*
 * StartJob retrieves & starts a given job
 */
func (s *Supervisor) StartJob(id int, wait bool) error {
	job, err := s.Mgr.GetJob(id)
	if err == nil {
		job.Start(wait)
	}
	return err
}

/*
 * StopJob retrieves & stops a given job
 */
func (s *Supervisor) StopJob(id int) error {
	job, err := s.Mgr.GetJob(id)
	if err == nil {
		job.Stop(true)
	}
	return err
}

/*
 * GetJob returns the job with the given id
 */
func (s *Supervisor) GetJob(id int) (*Job, error) {
	defer s.lock.Unlock()
	s.lock.Lock()
	return s.Mgr.GetJob(id)
}

/*
 * StartAllJobs starts all jobs & waits for start
 */
func (s *Supervisor) StartAllJobs(wait bool) {
	defer s.Mgr.lock.Unlock()
	s.Mgr.lock.Lock()
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

/*
 * StopAllJobs stops all jobs & waits for stop
 */
func (s *Supervisor) StopAllJobs(wait bool) {
	defer s.Mgr.lock.Unlock()
	s.Mgr.lock.Lock()
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

/*
 * HasJob returns number of jobs being managed
 */
func (s *Supervisor) HasJob(id int) bool {
	defer s.Mgr.lock.Unlock()
	s.Mgr.lock.Lock()
	_, ok := s.Mgr.Jobs[id]
	return ok
}

/*
 * ForAllJobs performs a callback on the managed jobs
 */
func (s *Supervisor) ForAllJobs(f func(job *Job)) {
	defer s.Mgr.lock.Unlock()
	s.Mgr.lock.Lock()
	for _, job := range s.Mgr.Jobs {
		f(job)
	}
}
