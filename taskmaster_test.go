package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	MockLogger()
	os.Exit(m.Run())
}

func TestStartStopSingle(t *testing.T) {
	file := "procfiles/StartStopSingle.yaml"
	ch := make(chan error)
	s := PrepareJobs(t, file)
	go func() {
		if err := s.StartJob(0); err != nil {
			ch <- err
		} else if err = s.StopJob(0); err != nil {
			ch <- err
		} else {
			ch <- nil
		}
	}()
	select {
	case err := <-ch:
		if err != nil {
			fmt.Println(err)
			t.Fail()
		}
		LogsContain(t, Buf.String(), []string{
			"Job 0 Successfully Started",
			"Sending Signal interrupt to Job 0",
			"Job 0 stopped with status: signal: interrupt",
			"Job 0 stopped by user, not restarting",
		})
	case <-time.After(time.Duration(5) * time.Second):
		t.Errorf("TestStartStopMulti timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestStartStopMulti(t *testing.T) {
	file := "procfiles/StartStopMulti.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		s.StartAllJobs()
		s.StopAllJobs()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 1 Successfully Started",
			"Sending Signal interrupt to Job 1",
			"Job 1 stopped with status: signal: interrupt",
			"Job 1 stopped by user, not restarting",
			"Job 2 Successfully Started",
			"Sending Signal interrupt to Job 2",
			"Job 2 stopped with status: signal: interrupt",
			"Job 2 stopped by user, not restarting",
		})
	case <-time.After(time.Duration(5) * time.Second):
		t.Errorf("TestStartStopMulti timed out, log:\n%s", Buf.String())
	}
}
func TestRestartAfterFailedStart(t *testing.T) {
	file := "procfiles/RestartAfterFailedStart.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		s.StartAllJobs()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		str := "Job 0 failed to start with error: fork\\/exec foo: no such file or directory"
		regex := regexp.MustCompile(str)
		matches := len(regex.FindAllString(logs, -1)) == 6
		matches = matches && strings.Contains(logs, "Creation of")
		matches = matches && len(strings.Split(logs, "\n")) == 7
		if !matches {
			t.Errorf(fmt.Sprintf("Error: Incorrect Logs:\n%s", logs))
		}
	case <-time.After(time.Duration(20) * time.Second):
		t.Errorf("TestRestartAfterFailedStart timed out, log:\n%s", Buf.String())
	}
}

func TestRestartAfterUnexpectedExit(t *testing.T) {
	file := "procfiles/RestartAfterUnexpectedExit.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		s.StartAllJobs()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		fmt.Println(Buf.String())
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRestartAfterFailedStart timed out, log:\n%s", Buf.String())
	}
}

func TestNoRestartAfterExpectedExit(t *testing.T) {
}

func TestNoRestartAfterExit(t *testing.T) {
}

func TestRestartAlways(t *testing.T) {
}

func TestStartTimeout(t *testing.T) {
}

func TestMultipleInstances(t *testing.T) {
}

func TestKillAfterIgnoredStopSignal(t *testing.T) {
}

func TestRedirectStdout(t *testing.T) {
}

func TestEnvVars(t *testing.T) {
}

func TestSetWorkingDir(t *testing.T) {
}

func TestUmask(t *testing.T) {
}
