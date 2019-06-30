package main

import (
	"io/ioutil"
	"strings"
	"syscall"

	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
)

// Log struct to print leveled logs
var Log *logging.Logger

// Exit messages for the various parsing errors
const (
	EXPECTEDEXITMSG = "Error: expectedExit must be signal name found in `man signal`"
	STOPSIGNALMSG   = "Error: stopSignal must be signal name found in `man signal`"
	INSTANCESMSG    = "Error: invalid instances value: %s\n"
	STARTCHECKUPMSG = "Error: invalid startCheckup value: %s\n"
	MAXRESTARTSMSG  = "Error: invalid maxrestarts value: %s\n"
	StopTimeoutMSG  = "Error: invalid StopTimeout value: %s\n"
	UMASKMSG        = "Error: invalid umask value: %s\n"
)

// JobConfig represents the config struct loaded from yaml
type JobConfig struct {
	ID            string
	Command       string
	Instances     string
	AtLaunch      string
	RestartPolicy string
	ExpectedExit  string
	StartCheckup  string
	MaxRestarts   string
	StopSignal    string
	StopTimeout   string
	EnvVars       string
	WorkingDir    string
	Umask         string
	pid           string
	Status        string
	Redirections  struct {
		Stdin  string
		Stdout string
		Stderr string
	}
}

// Check absence of errors, raise runtime error if found
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// SetDefaults translate JobConfig array into Job array, verifying inputs and setting defaults
func SetDefaults(configJobs []JobConfig) []Job {
	ids := make(map[int]bool)
	Jobs := []Job{}
	defaultUmask := syscall.Umask(0)

	syscall.Umask(defaultUmask)
	for _, cfg := range configJobs {
		var Job Job
		// an id to uniquely identify the Jobess
		Job.ParseID(&cfg, ids)
		// The command to use to launch the program
		Job.Args = strings.Fields(cfg.Command)
		// The number of Jobesses to start and keep running
		Job.ParseInt(&cfg, "Instances", 1, INSTANCESMSG)
		// Whether to start this program at launch or not
		Job.ParseAtLaunch(&cfg)
		// Whether the program should be restarted always, never, or on unexpected exits only
		Job.ParseRestartPolicy(&cfg)
		// Which return codes represent an "expected" exit Status
		Job.ParseInt(&cfg, "ExpectedExit", 0, EXPECTEDEXITMSG)
		// How long the program should be running after it’s started for it to be considered "successfully started"
		Job.ParseInt(&cfg, "StartCheckup", 1, STARTCHECKUPMSG)
		// How many times a restart should be attempted before aborting
		Job.ParseInt(&cfg, "MaxRestarts", 0, MAXRESTARTSMSG)
		// Which signal should be used to stop (i.e. exit gracefully) the program
		Job.StopSignal = Job.ParseSignal(&cfg, STOPSIGNALMSG)
		// How long to wait after a graceful stop before killing the program
		Job.ParseInt(&cfg, "StopTimeout", 1, StopTimeoutMSG)
		// Options to discard the program’s stdout/stderr or to redirect them to files
		Job.ParseRedirections(&cfg)
		// Environment variables to set before launching the program
		Job.EnvVars = strings.Fields(cfg.EnvVars)
		// A working directory to set before launching the program
		Job.WorkingDir = cfg.WorkingDir
		// An umask to set before launching the program
		Job.ParseInt(&cfg, "Umask", defaultUmask, UMASKMSG)
		Job.end = make(chan bool)
		// Add Job to array
		Jobs = append(Jobs, Job)
	}
	return Jobs
}

// LoadConfig opens config file and parses yaml syntax into array of JobConfig structs
func LoadConfig(args []string) ([]JobConfig, error) {
	configs := []JobConfig{}
	if buf, err := ioutil.ReadFile(args[0]); err != nil {
		return nil, err
	} else if err = yaml.Unmarshal(buf, &configs); err != nil {
		return nil, err
	}
	return configs, NewLogger(args[1:])
}
