package main

import (
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
	s := NewSupervisor("", NewManager())
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
	if len(new) != 1 || new[0].ID != 19 {
		t.Error("Error: new List should be [Job 19], actually", new)
	} else if len(changed) != 1 || changed[0].ID != 17 {
		t.Error("Error: changed List should be [Job 17], actually", changed)
	} else if len(old) != 1 || old[0].ID != 18 {
		t.Error("Error: old List should be [Job 18], actually", old)
	} else if len(current) != 1 || current[0].ID != 16 {
		t.Error("Error: current List should be [Job 16], actually", current)
	}
	Buf.Reset()
}
