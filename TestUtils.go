package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/op/go-logging"
)

var Buf bytes.Buffer

func MockLogger(out string) {
	var logOut io.Writer
	if out == "buf" {
		logOut = &Buf
	} else if out == "stdout" {
		logOut = os.Stdout
	}
	loggingBackend := logging.NewBackendFormatter(
		logging.NewLogBackend(logOut, "", 0),
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

//https:blog.sgmansfield.com/2015/12/goroutine-ids/
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
