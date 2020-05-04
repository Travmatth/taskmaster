package supervisor

import (
	"fmt"
	"sync"

	JOB "github.com/Travmatth/taskmaster/job"
)

//Manager controls access to jobs for the Supervisor struct
type Manager struct {
	Jobs map[int]*JOB.Job
	lock sync.Mutex
}

//NewManager create Manager struct
func NewManager() *Manager {
	return &Manager{
		Jobs: make(map[int]*JOB.Job),
	}
}

//AddSingleJob add single job
func (m *Manager) AddSingleJob(job *JOB.Job) {
	defer m.lock.Unlock()
	m.lock.Lock()
	m.Jobs[job.ID] = job
}

//AddMultiJobs adds multiple jobs
func (m *Manager) AddMultiJobs(jobs []*JOB.Job) {
	defer m.lock.Unlock()
	m.lock.Lock()
	for _, job := range jobs {
		m.Jobs[job.ID] = job
	}
}

//RemoveJob removes single job
func (m *Manager) RemoveJob(id int) *JOB.Job {
	defer m.lock.Unlock()
	m.lock.Lock()
	job := m.Jobs[id]
	delete(m.Jobs, id)
	return job
}

//GetJob retrieves job
func (m *Manager) GetJob(id int) (*JOB.Job, error) {
	defer m.lock.Unlock()
	m.lock.Lock()
	if job, ok := m.Jobs[id]; ok {
		return job, nil
	}
	return nil, fmt.Errorf("Manager Error: No Job with ID: %d", id)
}

//GetAllJobs retrieves all jobs
func (m *Manager) GetAllJobs(id int) []*JOB.Job {
	var jobs []*JOB.Job
	defer m.lock.Unlock()
	m.lock.Lock()
	for _, job := range m.Jobs {
		jobs = append(jobs, job)
	}
	return jobs
}
