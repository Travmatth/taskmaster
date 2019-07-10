package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/op/go-logging"
)

var buf bytes.Buffer

func mockLogger() {
	loggingBackend := logging.NewBackendFormatter(
		// logging.NewLogBackend(os.Stdout, "", 0),
		logging.NewLogBackend(&buf, "", 0),
		logging.MustStringFormatter(
			`[%{time:2006-01-02 15:04:05}] [%{level:.4s}] [%{shortfile}] - %{message}`,
		))
	Log, _ = logging.GetLogger("taskmaster")
	leveledBackend := logging.AddModuleLevel(loggingBackend)
	leveledBackend.SetLevel(logging.DEBUG, "")
	Log.SetBackend(leveledBackend)
}

func PrepareJobs(t *testing.T, file string) *Supervisor {
	buf.Reset()
	s := NewSupervisor("", NewManager())
	if buf, err := LoadFile(file); err != nil {
		panic(err)
	} else if configs, err := LoadJobs(buf); err != nil {
		panic(err)
	} else {
		jobs := SetDefaults(configs)
		s.Mgr.AddMultiJob(jobs)
		return s
	}
}

func LogsContain(t *testing.T, logStrings []string) {
	logs := buf.String()
	for _, str := range logStrings {
		if !strings.Contains(logs, str) {
			fmt.Println(logs)
			t.Errorf("Logs should contain: %s", str)
		}
	}
}

func TestMain(m *testing.M) {
	mockLogger()
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
		LogsContain(t, []string{
			"Job 0 Successfully Started",
			"Sending Signal interrupt to Job 0",
			"Job 0 stopped with status: signal: interrupt",
			"Job 0 stopped by user, not restarting",
		})
	case <-time.After(time.Duration(5) * time.Second):
		t.Errorf("TestStartStopMulti timed out, log:\n%s", buf.String())
	}
	buf.Reset()
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
		LogsContain(t, []string{
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
		t.Errorf("TestStartStopMulti timed out, log:\n%s", buf.String())
	}
}
func TestRestartAfterFailedStart(t *testing.T) {
	file := "procfiles/RestartAfterFailedStart.yaml"
	ch := make(chan struct{})
	s, err := PrepareJobs(t, file)
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}
	go func() {
		Log.Info("Starting TestRestartAfterFailedStart")
		s.StartAllJobs()
		s.StopAllJobs()
		Log.Info("Ending TestRestartAfterFailedStart")
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		fmt.Println(buf.String())
	case <-time.After(time.Duration(5) * time.Second):
		t.Errorf("TestRestartAfterFailedStart timed out, log:\n%s", buf.String())
	}
}

func TestRestartAfterUnexpectedExit(t *testing.T) {
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
