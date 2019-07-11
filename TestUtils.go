package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/op/go-logging"
)

var Buf bytes.Buffer

func MockLogger() {
	loggingBackend := logging.NewBackendFormatter(
		// logging.NewLogBackend(os.Stdout, "", 0),
		logging.NewLogBackend(&Buf, "", 0),
		logging.MustStringFormatter(
			`[%{time:2006-01-02 15:04:05}] [%{level:.4s}] [%{shortfile}] - %{message}`,
		))
	Log, _ = logging.GetLogger("taskmaster")
	leveledBackend := logging.AddModuleLevel(loggingBackend)
	leveledBackend.SetLevel(logging.DEBUG, "")
	Log.SetBackend(leveledBackend)
}

func PrepareJobs(t *testing.T, file string) *Supervisor {
	Buf.Reset()
	s := NewSupervisor("", NewManager())
	if Buf, err := LoadFile(file); err != nil {
		panic(err)
	} else if configs, err := LoadJobs(Buf); err != nil {
		panic(err)
	} else {
		jobs := SetDefaults(configs)
		s.Mgr.AddMultiJob(jobs)
		return s
	}
}

func LogsContain(t *testing.T, logs string, logStrings []string) {
	for _, str := range logStrings {
		if !strings.Contains(logs, str) {
			fmt.Println(logs)
			t.Errorf("Logs should contain: %s", str)
		}
	}
}
