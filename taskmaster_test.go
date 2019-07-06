package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/op/go-logging"
)

var buf bytes.Buffer
var Log *logging.Logger

func mockLogger() {
	loggingBackend := logging.NewBackendFormatter(
		logging.NewLogBackend(&buf, "", 0),
		logging.MustStringFormatter(
			`[%{time:2006-01-02 15:04:05}] [%{level:.4s}] [%{shortfile}] - %{message}`,
		))
	Log, _ = logging.GetLogger("taskmaster")
	leveledBackend := logging.AddModuleLevel(loggingBackend)
	leveledBackend.SetLevel(logging.DEBUG, "")
	Log.SetBackend(leveledBackend)
}

func parseJobs(buf []byte, t *testing.T) []*Job {
	configs, _ := LoadJobs(buf)
	return SetDefaults(configs, Log)
}

func init() {
	buf.Reset()
}

func TestMain(m *testing.M) {
	mockLogger()
	os.Exit(m.Run())
}

func TestStopSingle(t *testing.T) {
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
	s := NewSupervisor("", Log, NewManager(Log))
	jobs := parseJobs(file, t)
	s.Reload(jobs)
	s.lock.Lock()
	if p, ok := s.Mgr.Jobs[1]; !ok {
		s.Log.Debug("PID not available")
	} else if p.process == nil {
		s.Log.Debug("process not available")
	} else {
		s.Log.Debug("PID:", p.process.Pid)
	}
	s.lock.Unlock()
	time.Sleep(5)
	s.WaitForExit()
	time.Sleep(5)
	fmt.Println(buf.String())
}
