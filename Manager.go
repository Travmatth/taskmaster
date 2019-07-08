package main

import (
	"fmt"
	"sync"
)

type Manager struct {
	Jobs map[int]*Job
	lock sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		Jobs: make(map[int]*Job),
	}
}

func (m *Manager) AddJob(job *Job) {
	m.Jobs[job.ID] = job
}

func (m *Manager) RemoveJob(job *Job) {
	defer m.lock.Unlock()
	m.lock.Lock()
	delete(m.Jobs, job.ID)
}

func (m *Manager) RestartJob(job *Job) {
}

func (m *Manager) StartJob(job *Job) {
	job.Start(false)
}

func (m *Manager) StopAllJobs() {
	fmt.Println("Manager StopAllJobs called")
	var waitGroup sync.WaitGroup
	for _, job := range m.Jobs {
		m.lock.Lock()
		waitGroup.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			job.Stop(true)
		}(&waitGroup)
		m.lock.Unlock()
	}
	waitGroup.Wait()
}
