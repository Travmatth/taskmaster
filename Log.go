package main

import (
	"os"

	"github.com/op/go-logging"
)

func setLogLevel(args []string) logging.Level {
	switch args[0] {
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
func NewLogger(args []string) (*logging.Logger, error) {
	var f *os.File
	var err error
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if f, err = os.OpenFile(args[0], flags, 0666); err != nil {
		return nil, err
	}
	loggingBackend := logging.NewBackendFormatter(
		logging.NewLogBackend(f, "", 0),
		logging.MustStringFormatter(
			`[%{time:2006-01-02 15:04:05}] [%{level:.4s}] [%{shortfile}] - %{message}`,
		))
	Log, err := logging.GetLogger("taskmaster")
	leveledBackend := logging.AddModuleLevel(loggingBackend)
	leveledBackend.SetLevel(setLogLevel(args[1:]), "")
	Log.SetBackend(leveledBackend)
	return Log, err
}