package main

import (
	"flag"
	"os"
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
	s := PrepareSupervisor(t, file)
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
			t.Errorf("Err not nil:\n%s\n%s", err, Buf.String())
		} else {
			LogsContain(t, Buf.String(), []string{
				"Job 0 Instance 0 : Successfully Started after 1 second(s)",
				"Job 0 Instance 0 : Sending Signal interrupt",
				"Job 0 Instance 0 : exited with status: signal: interrupt",
				"Job 0 Instance 0 : stopped by user, not restarting",
			})
		}
	case <-time.After(time.Duration(5) * time.Second):
		t.Errorf("TestStartStopMulti timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestStartStopMultiJobs(t *testing.T) {
	file := "procfiles/StartStopMulti.yaml"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, file)
	go func() {
		s.StartAllJobs()
		s.StopAllJobs()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 1 Instance 0 : Successfully Started after 1 second(s)",
			"Job 2 Instance 0 : Successfully Started after 1 second(s)",
			"Job 1 Instance 0 : Sending Signal interrupt",
			"Job 2 Instance 0 : Sending Signal interrupt",
			"Job 1 Instance 0 : exited with status: signal: interrupt",
			"Job 2 Instance 0 : exited with status: signal: interrupt",
			"Job 1 Instance 0 : stopped by user, not restarting",
			"Job 2 Instance 0 : stopped by user, not restarting",
		})
	case <-time.After(time.Duration(5) * time.Second):
		t.Errorf("TestStartStopMulti timed out, log:\n%s", Buf.String())
	}
}
func TestRestartAfterFailedStart(t *testing.T) {
	file := "procfiles/RestartAfterFailedStart.yaml"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, file)
	go func() {
		s.StartAllJobs()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 3 Instance 0 : Creation of failed: failed to start with error: fork/exec foo: no such file or directory",
		})
	case <-time.After(time.Duration(20) * time.Second):
		t.Errorf("TestRestartAfterFailedStart timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestRestartAfterUnexpectedExit(t *testing.T) {
	file := "procfiles/RestartAfterUnexpectedExit.yaml"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(4)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		<-j.Instances[0].finishedCh
		<-j.Instances[0].finishedCh
		<-j.Instances[0].finishedCh
		<-j.Instances[0].finishedCh
		<-j.Instances[0].finishedCh
		time.Sleep(3)
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 4 Instance 0 : Successfully Started with no start checkup",
			"Job 4 Instance 0 : exited with status: exit status 1",
			"Job 4 Instance 0 : Encountered unexpected exit code 1 , restarting",
			"Job 4 Instance 0 : Successfully Started with no start checkup",
			"Job 4 Instance 0 : exited with status: exit status 1",
			"Job 4 Instance 0 : Encountered unexpected exit code 1 , restarting",
			"Job 4 Instance 0 : Successfully Started with no start checkup",
			"Job 4 Instance 0 : exited with status: exit status 1",
			"Job 4 Instance 0 : Encountered unexpected exit code 1 , restarting",
			"Job 4 Instance 0 : Successfully Started with no start checkup",
			"Job 4 Instance 0 : exited with status: exit status 1",
			"Job 4 Instance 0 : Encountered unexpected exit code 1 , restarting",
			"Job 4 Instance 0 : Successfully Started with no start checkup",
			"Job 4 Instance 0 : exited with status: exit status 1",
			"Job 4 Instance 0 : Encountered unexpected exit code 1 , restarting",
			"Job 4 Instance 0 : Successfully Started with no start checkup",
			"Job 4 Instance 0 : exited with status: exit status 0",
		})
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRestartAfterFailedStart timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestNoRestartAfterExpectedExit(t *testing.T) {
	file := "procfiles/NoRestartAfterExpectedExit.yaml"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(5)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 5 Instance 0 : Successfully Started with no start checkup",
			"Job 5 Instance 0 : exited with status: exit status 1",
		})
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestNoRestartAfterExpectedExit timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestNoRestartAfterExit(t *testing.T) {
	file := "procfiles/NoRestartAfterExit.yaml"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(6)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 6 Instance 0 : Successfully Started with no start checkup",
			"Job 6 Instance 0 : exited with status: exit status 1",
			"Job 6 Instance 0 : restart policy specifies do not restart",
		})
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestNoRestartAfterExit timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestRestartAlways(t *testing.T) {
	file := "procfiles/RestartAlways.yaml"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(7)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		<-j.Instances[0].finishedCh
		s.StopAllJobs()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 7 Instance 0 : Successfully Started with no start checkup",
			"Job 7 Instance 0 : exited with status: exit status 1",
			"Job 7 Instance 0 : Successfully Started with no start checkup",
			"Job 7 Instance 0 : exited with status: exit status 1",
			"Job 7 Instance 0 : Successfully Started with no start checkup",
			"Job 7 Instance 0 : Sending Signal interrupt",
			"Job 7 Instance 0 : exited with status: signal: interrupt",
			"Job 7 Instance 0 : stopped by user, not restarting",
		})
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestNoRestartAfterExit timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestStartTimeout(t *testing.T) {
	file := "procfiles/StartTimeout.yaml"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, file)
	go func() {
		j, _ := s.Mgr.GetJob(8)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		LogsContain(t, Buf.String(), []string{
			"Job 8 Instance 0 : exited with status: exit status 1",
			"Job 8 Instance 0 : monitor failed, program exit:  1  with job status 2",
			"Job 8 Instance 0 : Start failed, restarting",
			"Job 8 Instance 0 : Successfully Started after 2 second(s)",
			"Job 8 Instance 0 : exited with status: exit status 0",
			"Job 8 Instance 0 : restart policy specifies do not restart",
		})
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRestartAfterFailedStart timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestKillAfterIgnoredStopSignal(t *testing.T) {
	proc := "procfiles/KillAfterIgnoredStopSignal.yaml"
	test := "test/KillAfterIgnoredStopSignal.test"
	ch := make(chan error)
	s := PrepareSupervisor(t, proc)
	go func() {
		if err := s.StartJob(9); err != nil {
			ch <- err
		}
		if err := s.StopJob(9); err != nil {
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
			if contents, err := FileContains(test); err != nil {
				t.Errorf("Error: failed to open file with error %s", err)
			} else if contents != "INT caught" {
				t.Errorf("Error: incorrect file contents %s", contents)
			} else {
				LogsContain(t, Buf.String(), []string{
					"Job 9 Instance 0 : Sending Signal interrupt",
					"Job 9 Instance 0 : Successfully Started after 2 second(s)",
					"Job 9 Instance 0 : did not stop after timeout of  3 seconds SIGKILL issued",
					"Job 9 Instance 0 : exited with status: signal: killed",
					"Job 9 Instance 0 : stopped by user, not restarting",
				})
				os.Remove(test)
			}
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestStartStopMulti timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestRedirectStdout(t *testing.T) {
	procFile := "procfiles/RedirectStdout.yaml"
	testFile := "test/RedirectStdout.test"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, procFile)
	go func() {
		j, _ := s.Mgr.GetJob(10)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		if contents, err := FileContains(testFile); err != nil {
			t.Errorf("Error: file error\n%s\nlogs:%s", err, logs)
		} else if contents != "written to stdout" {
			t.Errorf("Error: incorrect string\n%s\nlogs:%s", contents, logs)
		} else {
			LogsContain(t, logs, []string{
				"Job 10 Instance 0 : Successfully Started with no start checkup",
				"Job 10 Instance 0 : exited with status: exit status 0",
				"Job 10 Instance 0 : restart policy specifies do not restart",
			})
			os.Remove(testFile)
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRedirectStdout timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestRedirectStderr(t *testing.T) {
	procFile := "procfiles/RedirectStderr.yaml"
	testFile := "test/RedirectStderr.test"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, procFile)
	go func() {
		j, _ := s.Mgr.GetJob(11)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		if contents, err := FileContains(testFile); err != nil {
			t.Errorf("Error: file error\n%s\nlogs:%s", err, logs)
		} else if contents != "written to stderr" {
			t.Errorf("Error: incorrect string\n%s\nlogs:%s", contents, logs)
		} else {
			LogsContain(t, logs, []string{
				"Job 11 Instance 0 : Successfully Started with no start checkup",
				"Job 11 Instance 0 : exited with status: exit status 0",
				"Job 11 Instance 0 : restart policy specifies do not restart",
			})
			os.Remove(testFile)
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRedirectStderr timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestEnvVars(t *testing.T) {
	procFile := "procfiles/EnvVars.yaml"
	testFile := "test/EnvVars.test"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, procFile)
	go func() {
		j, _ := s.Mgr.GetJob(12)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		if contents, err := FileContains(testFile); err != nil {
			t.Errorf("Error: file error\n%s\nlogs:%s", err, logs)
		} else if contents != "env vars working" {
			t.Errorf("Error: incorrect string\n%s\nlogs:%s", contents, logs)
		} else {
			LogsContain(t, logs, []string{
				"Job 12 Instance 0 : Successfully Started with no start checkup",
				"Job 12 Instance 0 : exited with status: exit status 0",
				"Job 12 Instance 0 : restart policy specifies do not restart",
			})
			os.Remove(testFile)
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestEnvVars timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestSetWorkingDir(t *testing.T) {
	procFile := "procfiles/SetWorkingDir.yaml"
	testFile := "test/SetWorkingDir.test"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, procFile)
	go func() {
		j, _ := s.Mgr.GetJob(13)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		time.Sleep(1)
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		if contents, err := FileContains(testFile); err != nil {
			t.Errorf("Error: file error\n%s\nlogs:%s", err, logs)
		} else if contents != "exists" {
			t.Errorf("Error: incorrect string\n%s\nlogs:%s", contents, logs)
		} else {
			LogsContain(t, logs, []string{
				"Job 13 Instance 0 : Successfully Started with no start checkup",
				"Job 13 Instance 0 : exited with status: exit status 0",
				"Job 13 Instance 0 : restart policy specifies do not restart",
			})
			os.Remove(testFile)
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRedirectStdout timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestUmask(t *testing.T) {
	procFile := "procfiles/SetUmask.yaml"
	testFile := "test/SetUmask.test"
	ch := make(chan struct{})
	s := PrepareSupervisor(t, procFile)
	go func() {
		j, _ := s.Mgr.GetJob(14)
		s.StartAllJobs()
		<-j.Instances[0].finishedCh
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		logs := Buf.String()
		if contents, err := FileContains(testFile); err != nil {
			t.Errorf("Error: file error\n%s\nlogs:%s", err, logs)
		} else if contents != "0000" {
			t.Errorf("Error: incorrect string\n%s\nlogs:\n%s", contents, logs)
		} else {
			LogsContain(t, logs, []string{
				"Job 14 Instance 0 : Successfully Started with no start checkup",
				"Job 14 Instance 0 : exited with status: exit status 0",
				"Job 14 Instance 0 : restart policy specifies do not restart",
			})
			os.Remove(testFile)
		}
	case <-time.After(time.Duration(10) * time.Second):
		t.Errorf("TestRedirectStdout timed out, logs:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestStartStopMultipleInstances(t *testing.T) {
	procFile := "procfiles/StartStopMultipleInstances.yaml"
	ch := make(chan error)
	s := PrepareSupervisor(t, procFile)
	go func() {
		if err := s.StartJob(15); err != nil {
			ch <- err
		} else if err = s.StopJob(15); err != nil {
			ch <- err
		} else {
			ch <- nil
		}
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Errorf("Err not nil:\n%s\nLogs:\n%s\n", Buf.String(), err)
		} else {
			LogsContain(t, Buf.String(), []string{
				"Job 15 Instance 0 : Successfully Started after 1 second(s)",
				"Job 15 Instance 1 : Successfully Started after 1 second(s)",
				"Job 15 Instance 0 : Sending Signal interrupt",
				"Job 15 Instance 1 : Sending Signal interrupt",
				"Job 15 Instance 0 : exited with status: signal: interrupt",
				"Job 15 Instance 1 : exited with status: signal: interrupt",
				"Job 15 Instance 0 : stopped by user, not restarting",
				"Job 15 Instance 1 : stopped by user, not restarting",
			})
		}
	case <-time.After(time.Duration(5) * time.Second):
		t.Errorf("TestStartStopMulti timed out, log:\n%s", Buf.String())
	}
	Buf.Reset()
}

func TestChildProcessesKilled(t *testing.T) {
	Buf.Reset()
}
