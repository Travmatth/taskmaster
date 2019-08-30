package main

import "fmt"

type Job struct {
	ID        int
	Instances []*Instance
	pool      int
	cfg       *JobConfig
	AtLaunch  bool
}

func (j *Job) Start(wait bool) {
	for _, instance := range j.Instances {
		instance.StartInstance(wait)
	}
}

func (j *Job) Stop(wait bool) {
	for _, instance := range j.Instances {
		instance.StopInstance(wait)
	}
}

func (j Job) String() string {
	return fmt.Sprintf("Job %d", j.ID)
}
