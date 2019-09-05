package main

import (
	"os"
	"testing"
)

func processJobsFromFiles(files ...string) []*Job {
	var total []*Job
	for _, f := range files {
		jobs, _ := LoadJobsFromFile(f)
		total = append(total, jobs...)
	}
	return total
}

func TestSupervisorDiffJobs(t *testing.T) {
	ch := make(chan os.Signal)
	s := NewSupervisor("", NewManager(), ch, "")
	orig := processJobsFromFiles(
		"procfiles/DiffCurrentJobs.yaml",
		"procfiles/DiffOldJobs.yaml",
	)
	s.AddMultiJobs(orig)
	next := processJobsFromFiles(
		"procfiles/DiffCurrentJobs.yaml",
		"procfiles/DiffNewJobs.yaml",
		"procfiles/DiffChangedJobs.yaml",
	)
	next = append(next[:1], next[2:]...)
	current, old, changed, new := s.DiffJobs(next)
	if len(new) != 2 || (new[0].ID != 19 && new[0].ID != 17) {
		t.Error("Error: new List should be [Job 19, Job 17], actually", new)
	} else if len(changed) != 1 || changed[0].ID != 17 {
		t.Error("Error: changed List should be [Job 17], actually", changed)
	} else if len(old) != 1 || old[0].ID != 18 {
		t.Error("Error: old List should be [Job 18], actually", old)
	} else if len(current) != 1 || current[0].ID != 16 {
		t.Error("Error: current List should be [Job 16], actually", current)
	} else if len(s.Mgr.Jobs) != 0 {
		t.Errorf("Error: DiffJobs Should remove all jobs from Manager")
	}
	Buf.Reset()
}

func TestSupervisorReloadJobs(t *testing.T) {
	ch := make(chan os.Signal)
	s := NewSupervisor("", NewManager(), ch, "")
	orig := processJobsFromFiles(
		"procfiles/DiffCurrentJobs.yaml",
		"procfiles/DiffOldJobs.yaml",
	)
	s.AddMultiJobs(orig)
	next := processJobsFromFiles(
		"procfiles/DiffCurrentJobs.yaml",
		"procfiles/DiffNewJobs.yaml",
		"procfiles/DiffChangedJobs.yaml",
	)
	next = append(next[:1], next[2:]...)
	s.StartAllJobs(true)
	s.Reload(next, true)
	LogsContain(t, Buf.String(), []string{
		"Job 18 Instance 0 : Successfully Started with no start checkup",
		"Job 17 Instance 0 : Successfully Started with no start checkup",
		"Job 16 Instance 0 : Successfully Started with no start checkup",
		"Supervisor diffing next jobs: current Job 16",
		"Supervisor diffing next jobs: new Job 19",
		"Supervisor diffing next jobs: changed Job 17",
		"Job 18 Instance 0 : Sending Signal interrupt",
		"Job 18 Instance 0 : exited with status: signal: interrupt",
		"Job 18 Instance 0 : stopped by user, not restarting",
		"Job 17 Instance 0 : Sending Signal interrupt",
		"Job 17 Instance 0 : exited with status: signal: interrupt",
		"Job 17 Instance 0 : stopped by user, not restarting",
		"Job 19 Instance 0 : Successfully Started with no start checkup",
		"Job 17 Instance 0 : Successfully Started with no start checkup",
		"Job 17 Instance 1 : Successfully Started with no start checkup",
	})
	Buf.Reset()
}

func TestSupervisorReloadJobsScaleUp(t *testing.T) {
	s := PrepareSupervisor(t, "procfiles/DiffCurrentJobs.yaml")
	s.StartAllJobs(true)
	job, _ := s.GetJob(17)
	orig := job.Instances[0].StartTime
	next, _ := LoadJobsFromFile("procfiles/DiffChangedJobs.yaml")
	s.Reload(next, true)
	LogsContain(t, Buf.String(), []string{
		"Job 17 Instance 0 : Successfully Started with no start checkup",
		"Job 16 Instance 0 : Successfully Started with no start checkup",
		"Supervisor diffing next jobs: changed Job 17",
		"Job 16 Instance 0 : Sending Signal interrupt",
		"Job 16 Instance 0 : exited with status: signal: interrupt",
		"Job 16 Instance 0 : stopped by user, not restarting",
		"Job 17 Instance 1 : Successfully Started with no start checkup",
	})
	newJob, _ := s.GetJob(17)
	nextTime := newJob.Instances[0].StartTime
	if !orig.Equal(nextTime) {
		t.Errorf("Error: Supervisor shouldn't replace running instances when scaling up")
	}
	Buf.Reset()
}

func TestSupervisorReloadJobsScaleDown(t *testing.T) {
	s := PrepareSupervisor(t, "procfiles/DiffChangedJobs.yaml")
	s.StartAllJobs(true)
	job, _ := s.GetJob(17)
	orig := job.Instances[0].StartTime
	next, _ := LoadJobsFromFile("procfiles/DiffCurrentJobs.yaml")
	s.Reload(next, true)
	LogsContain(t, Buf.String(), []string{
		"Job 17 Instance 0 : Successfully Started with no start checkup",
		"Job 17 Instance 1 : Successfully Started with no start checkup",
		"Supervisor diffing next jobs: new Job 16",
		"Supervisor diffing next jobs: changed Job 17",
		"Job 17 Instance 1 : Sending Signal interrupt",
		"Job 17 Instance 1 : exited with status: signal: interrupt",
		"Job 17 Instance 1 : stopped by user, not restarting",
		"Job 16 Instance 0 : Successfully Started with no start checkup",
	})
	newJob, _ := s.GetJob(17)
	nextTime := newJob.Instances[0].StartTime
	if !orig.Equal(nextTime) {
		t.Errorf("Error: Supervisor shouldn't replace running instances when scaling down")
	}
	Buf.Reset()
}
