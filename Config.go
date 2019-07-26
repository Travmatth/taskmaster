package main

import (
	"fmt"
	"io/ioutil"
	"os"
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
	instances, err := strconv.Atoi(cfg.Instances)
	if err != nil && instances != "" {
		fmt.Print("Job", cfg)
		fmt.Printf(INSTANCESMSG, cfg.Instances)
		os.Exit(1)
	} else if instances == "" {
		instances = 1
	}
	job.Instances = make([]*Instance, instances)
	job.pool = instances
	for _, instance := range job.Instances {
		// an id to uniquely identify the Jobess
		instance.ParseID(cfg, ids)
		// The command to use to launch the program
		instance.args = strings.Fields(cfg.Command)
		// The number of Jobesses to start and keep running
		instance.ParseInt(cfg, "Instances", 1, INSTANCESMSG)
		// Whether to start this program at launch or not
		instance.ParseAtLaunch(cfg)
		// Whether the program should be restarted always, never, or on unexpected exits only
		instance.ParserestartPolicy(cfg)
		// Which return codes represent an "expected" exit Status
		instance.ParseInt(cfg, "ExpectedExit", 0, EXPECTEDEXITMSG)
		// How long the program should be running after it’s started for it to be considered "successfully started"
		instance.ParseInt(cfg, "StartCheckup", 1, STARTCHECKUPMSG)
		// How many times a restart should be attempted before aborting
		// instance.ParseInt(cfg, "MaxRestarts", 0, MAXRESTARTSMSG)
		if max, err := strconv.Atoi(cfg.MaxRestarts); err == nil {
			instance.MaxRestarts = int32(max)
		} else {
			instance.MaxRestarts = 0
		}
		instance.Restarts = new(int32)
		// Which signal should be used to stop (i.e. exit gracefully) the program
		instance.StopSignal = instance.ParseSignal(cfg, STOPSIGNALMSG)
		// How long to wait after a graceful stop before killing the program
		// instance.ParseInt(cfg, "StopTimeout", 1, STOPTIMEOUTMSG)
		if timeout, err := strconv.Atoi(cfg.StopTimeout); err == nil {
			instance.StopTimeout = timeout
		} else {
			instance.StopTimeout = 5
		}
		// Options to discard the program’s stdout/stderr or to redirect them to files
		instance.ParseRedirections(cfg)
		// Environment variables to set before launching the program
		instance.EnvVars = strings.Fields(cfg.EnvVars)
		// A working directory to set before launching the program
		instance.WorkingDir = cfg.WorkingDir
		// An umask to set before launching the program
		instance.ParseInt(cfg, "Umask", defaultUmask, UMASKMSG)
		// Add conditional var to struct
		instance.condition = sync.NewCond(&instance.mutex)
		instance.finishedCh = make(chan struct{})
		// Add config to job struct
		instance.cfg = cfg
	}
	return job
}

func GetDefaultUmask() int {
	defaultUmask := syscall.Umask(0)
	defer syscall.Umask(defaultUmask)
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

// LoadConfig opens config file and parses yaml syntax into array of JobConfig structs
func LoadJobs(buf []byte) ([]JobConfig, error) {
	configs := []JobConfig{}
	if err := yaml.Unmarshal(buf, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}
