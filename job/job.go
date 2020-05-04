package job

import (
	"fmt"

	CFG "github.com/Travmatth/taskmaster/config"
	INST "github.com/Travmatth/taskmaster/instance"
)

type Job struct {
	ID        int
	Instances []*INST.Instance
	Pool      int
	Cfg       *CFG.JobConfig
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
