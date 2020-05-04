package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/Travmatth/taskmaster/log"
	"github.com/op/go-logging"
)

var Buf bytes.Buffer

func FileContains(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func MockLogger(out string) {
	var logOut io.Writer
	if out == "buf" {
		logOut = &Buf
	} else if out == "stdout" {
		logOut = os.Stdout
	}
	loggingBackend := logging.NewBackendFormatter(
		logging.NewLogBackend(logOut, "", 0),
		logging.MustStringFormatter(`%{message}`))
	Log, _ = logging.GetLogger("taskmaster")
	leveledBackend := logging.AddModuleLevel(loggingBackend)
	leveledBackend.SetLevel(logging.DEBUG, "")
	Log.SetBackend(leveledBackend)
}

func LogsContain(t *testing.T, logs string, logStrings []string) {
	ok := true
	logs = logs + "\n"
	fullLogs := logs
	for _, str := range logStrings {
		if !strings.Contains(logs, str) {
			ok = false
			break
		}
		logs = strings.Replace(logs, str, "", 1)
	}
	logs = strings.ReplaceAll(logs, "\n", "")
	if !ok || logs != "" {
		message := "Log Error: Logs should contain:\n%s\nContains:\n%s\n"
		t.Errorf(message, strings.Join(logStrings, "\n"), fullLogs)
	}
}
