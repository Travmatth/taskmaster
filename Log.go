package main

import (
	"io"
	"os"
	"strings"

	"github.com/op/go-logging"
)

var Log *logging.Logger

func setLogLevel(level string) logging.Level {
	switch strings.ToUpper(level) {
	case "CRITICAL":
		return logging.CRITICAL
	case "ERROR":
		return logging.ERROR
	case "WARNING":
		return logging.WARNING
	case "NOTICE":
		return logging.NOTICE
	case "INFO":
		return logging.INFO
	case "DEBUG":
		return logging.DEBUG
	default:
		return logging.INFO
	}
}

// NewLogger creates logger for use in program
func NewLogger(name string, level string) error {
	var f *os.File
	var out io.Writer
	var err error

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if name == strings.ToLower("stdout") {
		out = os.Stdout
	} else if f, err = os.OpenFile(name, flags, 0666); err != nil {
		return err
	} else {
		out = f
	}
	loggingBackend := logging.NewBackendFormatter(
		logging.NewLogBackend(out, "", 0),
		logging.MustStringFormatter(
			`[%{time:2006-01-02 15:04:05}] [%{level:.4s}] [%{shortfile}] - %{message}`,
		))
	Log, err = logging.GetLogger("taskmaster")
	leveledBackend := logging.AddModuleLevel(loggingBackend)
	leveledBackend.SetLevel(setLogLevel(level), "")
	Log.SetBackend(leveledBackend)
	return err
}
