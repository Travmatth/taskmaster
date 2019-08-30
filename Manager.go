package main

import (
	"fmt"
	"sync"
)

type Manager struct {
	Jobs map[int]*Job
	lock sync.Mutex
}

//NewManager create Manager struct
func NewManager() *Manager {
	return &Manager{
		Jobs: make(map[int]*Job),
	}
}

//AddSingleJob add single job
func (m *Manager) AddSingleJob(job *Job) {
	defer m.lock.Unlock()
	m.lock.Lock()
	m.Jobs[job.ID] = job
}

//AddMultiJobs adds multiple jobs
func (m *Manager) AddMultiJobs(jobs []*Job) {
	defer m.lock.Unlock()
	m.lock.Lock()
	for _, job := range jobs {
		m.Jobs[job.ID] = job
	}
}

//RemoveJob removes single job
func (m *Manager) RemoveJob(id int) *Job {
	defer m.lock.Unlock()
	m.lock.Lock()
	job := m.Jobs[id]
	delete(m.Jobs, id)
	return job
}

//GetJob retrieves job
func (m *Manager) GetJob(id int) (*Job, error) {
	defer m.lock.Unlock()
	m.lock.Lock()
	if job, ok := m.Jobs[id]; ok {
		return job, nil
	}
	return nil, fmt.Errorf("Manager Error: No Job with ID: %d", id)
}

//GetAllJobs retrieves all jobs
func (m *Manager) GetAllJobs(id int) []*Job {
	var jobs []*Job
	defer m.lock.Unlock()
	m.lock.Lock()
	for _, job := range m.Jobs {
		jobs = append(jobs, job)
	}
	return jobs
}
