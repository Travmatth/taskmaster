package main

import (
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"gopkg.in/yaml.v2"
)

// Exit messages for the various parsing errors
const (
	EXPECTEDEXITMSG = "Error: expectedExit must be signal name found in `man signal`"
	STOPSIGNALMSG   = "Error: stopSignal must be signal name found in `man signal`"
	INSTANCESMSG    = "Error: invalid instances value: %s\n"
	STARTCHECKUPMSG = "Error: invalid startCheckup value: %s\n"
	MAXRESTARTSMSG  = "Error: invalid maxrestarts value: %s\n"
	STOPTIMEOUTMSG  = "Error: invalid StopTimeout value: %s\n"
	UMASKMSG        = "Error: invalid umask value: %s\n"
)

// JobConfig represents the config struct loaded from yaml
type JobConfig struct {
	ID            string `json:"ID"`
	Command       string `json:"Command"`
	Instances     string `json:"Instances"`
	AtLaunch      string `json:"AtLaunch"`
	RestartPolicy string `json:"RestartPolicy"`
	ExpectedExit  string `json:"ExpectedExit"`
	StartCheckup  string `json:"StartCheckup"`
	MaxRestarts   string `json:"MaxRestarts"`
	StopSignal    string `json:"StopSignal"`
	StopTimeout   string `json:"StopTimeout"`
	EnvVars       string `json:"EnvVars"`
	WorkingDir    string `json:"WorkingDir"`
	Umask         string `json:"Umask"`
	Redirections  struct {
		Stdin  string `json:"Stdin"`
		Stdout string `json:"Stdout"`
		Stderr string `json:"Stderr"`
	}
}

// Check absence of errors, raise runtime error if found
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func SetDefault(cfg *JobConfig, job *Job, ids map[int]bool, defaultUmask int) *Job {
	// an id to uniquely identify the Jobess
	job.ParseID(cfg, ids)
	// The command to use to launch the program
	job.args = strings.Fields(cfg.Command)
	// The number of Jobesses to start and keep running
	job.ParseInt(cfg, "Instances", 1, INSTANCESMSG)
	// Whether to start this program at launch or not
	job.ParseAtLaunch(cfg)
	// Whether the program should be restarted always, never, or on unexpected exits only
	job.ParserestartPolicy(cfg)
	// Which return codes represent an "expected" exit Status
	job.ParseInt(cfg, "ExpectedExit", 0, EXPECTEDEXITMSG)
	// How long the program should be running after it’s started for it to be considered "successfully started"
	job.ParseInt(cfg, "StartCheckup", 1, STARTCHECKUPMSG)
	// How many times a restart should be attempted before aborting
	// job.ParseInt(cfg, "MaxRestarts", 0, MAXRESTARTSMSG)
	if max, err := strconv.Atoi(cfg.MaxRestarts); err == nil {
		job.MaxRestarts = int32(max)
	} else {
		job.MaxRestarts = 0
	}
	job.Restarts = new(int32)
	// Which signal should be used to stop (i.e. exit gracefully) the program
	job.StopSignal = job.ParseSignal(cfg, STOPSIGNALMSG)
	// How long to wait after a graceful stop before killing the program
	// job.ParseInt(cfg, "StopTimeout", 1, STOPTIMEOUTMSG)
	if timeout, err := strconv.Atoi(cfg.StopTimeout); err == nil {
		job.StopTimeout = timeout
	} else {
		job.StopTimeout = 5
	}
	// Options to discard the program’s stdout/stderr or to redirect them to files
	job.ParseRedirections(cfg)
	// Environment variables to set before launching the program
	job.EnvVars = strings.Fields(cfg.EnvVars)
	// A working directory to set before launching the program
	job.WorkingDir = cfg.WorkingDir
	// An umask to set before launching the program
	job.ParseInt(cfg, "Umask", defaultUmask, UMASKMSG)
	// Add conditional var to struct
	job.condition = sync.NewCond(&job.mutex)
	job.finishedCh = make(chan struct{})
	// Add config to job struct
	job.cfg = cfg
	return job
}

func GetDefaultUmask() int {
	defaultUmask := syscall.Umask(0)
	syscall.Umask(defaultUmask)
	return defaultUmask
}

// SetDefaults translate JobConfig array into Job array, verifying inputs and setting defaults
func SetDefaults(configJobs []JobConfig) []*Job {
	ids := make(map[int]bool)
	Jobs := []*Job{}
	umask := GetDefaultUmask()

	for _, cfg := range configJobs {
		var job Job
		Jobs = append(Jobs, SetDefault(&cfg, &job, ids, umask))
	}
	return Jobs
}

//LoadFile Reads given Procfile into buffer
func LoadFile(file string) ([]byte, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// return configs, NewLogger(args[1:])
// LoadConfig opens config file and parses yaml syntax into array of JobConfig structs
func LoadJobs(buf []byte) ([]JobConfig, error) {
	configs := []JobConfig{}
	if err := yaml.Unmarshal(buf, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}
