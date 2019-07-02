package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/op/go-logging"
)

var s *Supervisor
var buf bytes.Buffer

func mockLogger() (*logging.Logger, error) {
	loggingBackend := logging.NewBackendFormatter(
		logging.NewLogBackend(&buf, "", 0),
		logging.MustStringFormatter(
			`[%{time:2006-01-02 15:04:05}] [%{level:.4s}] [%{shortfile}] - %{message}`,
		))
	Log, err := logging.GetLogger("taskmaster")
	leveledBackend := logging.AddModuleLevel(loggingBackend)
	leveledBackend.SetLevel(logging.DEBUG, "")
	Log.SetBackend(leveledBackend)
	return Log, err
}

func parseProcs(buf []byte, t *testing.T) []*Job {
	t.Helper()
	configProcs, _ := LoadJobs(buf)
	return SetDefaults(configProcs)
}

func init() {
	buf.Reset()
}

func TestMain(m *testing.M) {
	Log, _ := mockLogger()
	s = NewSupervisor(Log)
	os.Exit(m.Run())
}

func TestStopSingle(t *testing.T) {
	timeoutCh, funcCh := make(chan bool), make(chan bool)
	var file = []byte(`
- id: 1
  command: /bin/sleep 9999
  instances: 1
  atLaunch: true
  restartpolicy: always
  expectedexit: 0
  startcheckup: 1
  maxrestarts: 0
  stopsignal: SIGINT
  StopTimeout: 10
  redirections:
    stdin:
    stdout:
    stderr:
  envvars:
  workingdir:
  umask:
`)
	jobs := parseProcs(file, t)
	go func() {
		finished := make(chan struct{})
		go func() {
			s.StartAllJobs(jobs)
			finished <- struct{}{}
		}()
		s.stopCh <- 1
		<-finished
		funcCh <- true
	}()
	go func() {
		time.Sleep(time.Duration(5) * time.Second)
		timeoutCh <- true
	}()
	select {
	case <-funcCh:
		fmt.Println("TestStopSingle:\n", buf.String())
	case <-timeoutCh:
		t.Errorf(fmt.Sprintf("Timeout failed:\n%s", buf.String()))
	}
}
