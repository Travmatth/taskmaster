package main

type Job struct {
	ID        int
	Instances []*Instance
	pool      int
}

func (j *Job) Start(wait bool) {
	for _, instance := range j.Instances {
		instance.start(wait)
	}
}

func (j *Job) Stop(wait bool) {
	for _, instance := range j.Instances {
		instance.StopInstance(wait)
	}
}
