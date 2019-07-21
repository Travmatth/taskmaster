package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	var logOut string
	flag.StringVar(&logOut, "logs", "buf", "Log file output")
	flag.Parse()
	MockLogger(logOut)
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
			t.Errorf("Err not nil:\n%s", Buf.String())
		} else {
			LogsContain(t, Buf.String(), []string{
				"Job 0 Successfully Started",
				"Sending Signal interrupt to Job 0",
				"Job 0 exited with status: signal: interrupt",
				"Job 0 stopped by user, not restarting",
			})
		}
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
			"Job 1 exited with status: signal: interrupt",
			"Job 1 stopped by user, not restarting",
			"Job 2 Successfully Started",
			"Sending Signal interrupt to Job 2",
			"Job 2 exited with status: signal: interrupt",
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
	Buf.Reset()
}

func TestRestartAfterUnexpectedExit(t *testing.T) {
	file := "procfiles/RestartAfterUnexpectedExit.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(0)
		s.StartAllJobs()
		<-j.finishedCh
		<-j.finishedCh
		<-j.finishedCh
		<-j.finishedCh
		<-j.finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 0 Successfully Started",
			"Job 0 exited with status: exit status 1",
			"Job 0 Encountered unexpected exit code 1 , restarting",
			"Job 0 Successfully Started",
			"Job 0 exited with status: exit status 1",
			"Job 0 Encountered unexpected exit code 1 , restarting",
			"Job 0 Successfully Started",
			"Job 0 exited with status: exit status 1",
			"Job 0 Encountered unexpected exit code 1 , restarting",
			"Job 0 Successfully Started",
			"Job 0 exited with status: exit status 1",
			"Job 0 Encountered unexpected exit code 1 , restarting",
			"Job 0 Successfully Started",
			"Job 0 exited with status: exit status 1",
			"Job 0 Encountered unexpected exit code 1 , restarting",
		})
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRestartAfterFailedStart timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestNoRestartAfterExpectedExit(t *testing.T) {
	file := "procfiles/NoRestartAfterExpectedExit.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(0)
		s.StartAllJobs()
		<-j.finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		str := "Job 0 Encountered unexpected exit code 1 , restarting"
		if strings.Contains(logs, str) {
			t.Errorf(fmt.Sprintf("Error: Incorrect Logs:\n%s", logs))
		} else {
			LogsContain(t, Buf.String(), []string{
				"Job 0 Successfully Started",
				"Job 0 exited with status: exit status 1",
			})
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestNoRestartAfterExpectedExit timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestNoRestartAfterExit(t *testing.T) {
	file := "procfiles/NoRestartAfterExit.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(0)
		s.StartAllJobs()
		<-j.finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		str := "Job 0 Encountered unexpected exit code 1 , restarting"
		if strings.Count(logs, str) > 1 {
			t.Errorf(fmt.Sprintf("Error: Incorrect Logs:\n%s", logs))
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestNoRestartAfterExit timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestRestartAlways(t *testing.T) {
	file := "procfiles/RestartAlways.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(0)
		s.StartAllJobs()
		<-j.finishedCh
		<-j.finishedCh
		s.StopAllJobs()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 0 Successfully Started",
			"Job 0 exited with status: exit status 1",
			"Job 0 Successfully Started",
			"Job 0 exited with status: exit status 1",
			"Job 0 Successfully Started",
			"Sending Signal interrupt to Job 0",
			"Job 0 exited with status: signal: interrupt",
			"Job 0 stopped by user, not restarting",
		})
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestNoRestartAfterExit timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestStartTimeout(t *testing.T) {
	file := "procfiles/StartTimeout.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		s.StartAllJobs()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 0 exited with status: exit status 1",
			"Job 0 monitor failed, program exit:  1  with job status 2",
			"Job 0 Start failed, restarting",
			"Job 0 Successfully Started",
		})
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRestartAfterFailedStart timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestKillAfterIgnoredStopSignal(t *testing.T) {
	file := "procfiles/KillAfterIgnoredStopSignal.yaml"
	ch := make(chan error)
	s := PrepareJobs(t, file)
	go func() {
		// j, _ := s.Mgr.GetJob(0)
		if err := s.StartJob(0); err != nil {
			ch <- err
		}
		// j.Redirections[1].WriteString("Redirections working")
		if err := s.StopJob(0); err != nil {
			ch <- err
		} else {
			ch <- nil
		}
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Errorf("Err not nil:\n%s", Buf.String())
		} else {
			if contents, err := FileContains("test/KillAfterIgnoredStopSignal.test"); err != nil {
				t.Errorf("Error: failed to open file with error %s", err)
			} else if contents != "INT caught" {
				t.Errorf("Error: incorrect file contents %s", contents)
			} else {
				LogsContain(t, Buf.String(), []string{
					"Job 0 Successfully Started",
					"Sending Signal interrupt to Job 0",
					"Job 0 did not stop after timeout of  3 seconds SIGKILL issued",
					"Job 0 exited with status: signal: killed",
					"Job 0 stopped by user, not restarting",
				})
			}
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestStartStopMulti timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestRedirectStdout(t *testing.T) {
	file := "procfiles/RedirectStdout.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(0)
		s.StartAllJobs()
		<-j.finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		if file, err := FileContains("test/RedirectStdout.test"); err != nil {
			t.Errorf("Error: file error\n%s\nlogs:%s", err, logs)
		} else if file != "written to stdout" {
			t.Errorf("Error: incorrect string\n%s\nlogs:%s", file, logs)
		} else {
			LogsContain(t, logs, []string{
				"Job 0 Successfully Started",
				"Job 0 exited with status: exit status 0",
				"Job 0 restart policy specifies do not restart",
			})
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRedirectStdout timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestRedirectStderr(t *testing.T) {
	file := "procfiles/RedirectStderr.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(0)
		s.StartAllJobs()
		<-j.finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		if file, err := FileContains("test/RedirectStderr.test"); err != nil {
			t.Errorf("Error: file error\n%s\nlogs:%s", err, logs)
		} else if file != "written to stderr" {
			t.Errorf("Error: incorrect string\n%s\nlogs:%s", file, logs)
		} else {
			LogsContain(t, logs, []string{
				"Job 0 Successfully Started",
				"Job 0 exited with status: exit status 0",
				"Job 0 restart policy specifies do not restart",
			})
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRedirectStderr timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestEnvVars(t *testing.T) {
	file := "procfiles/EnvVars.yaml"
	ch := make(chan struct{})
	s := PrepareJobs(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(0)
		s.StartAllJobs()
		<-j.finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		if file, err := FileContains("test/EnvVars.test"); err != nil {
			t.Errorf("Error: file error\n%s\nlogs:%s", err, logs)
		} else if file != "env vars working" {
			t.Errorf("Error: incorrect string\n%s\nlogs:%s", file, logs)
		} else {
			LogsContain(t, logs, []string{
				"Job 0 Successfully Started",
				"Job 0 exited with status: exit status 0",
				"Job 0 restart policy specifies do not restart",
			})
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestEnvVars timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestMultipleInstances(t *testing.T) {
	Buf.Reset()
}

func TestSetWorkingDir(t *testing.T) {
	Buf.Reset()
}

func TestUmask(t *testing.T) {
	Buf.Reset()
}
