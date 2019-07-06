package main

import (
	"fmt"
	"sync"

	"github.com/op/go-logging"
)

type Manager struct {
	Jobs map[int]*Job
	lock sync.Mutex
	Log  *logging.Logger
}

func NewManager(log *logging.Logger) *Manager {
	return &Manager{
		Jobs: make(map[int]*Job),
		Log:  log,
	}
}

func (m *Manager) AddJob(job *Job) {
	defer m.lock.Unlock()
	m.lock.Lock()
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
