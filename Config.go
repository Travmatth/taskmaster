package main

import (
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
	// an id to uniquely identify the Jobs
	job.ParseID(cfg, ids)
	if err != nil && cfg.Instances != "" {
		os.Exit(1)
	} else if cfg.Instances == "" {
		instances = 1
	}
	job.Instances = make([]*Instance, instances)
	job.pool = instances
	// Add config to job struct
	job.cfg = cfg
	// Whether to start this program at launch or not
	job.ParseAtLaunch(cfg)
	for i := 0; i < instances; i++ {
		var inst Instance
		job.Instances[i] = &inst
		// ID to uniquely identify the Job & instance
		inst.JobID = job.ID
		inst.InstanceID = i
		// The command to use to launch the program
		inst.args = strings.Fields(cfg.Command)
		// The number of Jobesses to start and keep running
		inst.ParseInt(cfg, "Instances", 1, INSTANCESMSG)
		// Whether the program should be restarted always, never, or on unexpected exits only
		inst.ParserestartPolicy(cfg)
		// Which return codes represent an "expected" exit Status
		inst.ParseInt(cfg, "ExpectedExit", 0, EXPECTEDEXITMSG)
		// How long the program should be running after it’s started for it to be considered "successfully started"
		inst.ParseInt(cfg, "StartCheckup", 1, STARTCHECKUPMSG)
		// How many times a restart should be attempted before aborting
		// inst.ParseInt(cfg, "MaxRestarts", 0, MAXRESTARTSMSG)
		if max, err := strconv.Atoi(cfg.MaxRestarts); err == nil {
			inst.MaxRestarts = int32(max)
		} else {
			inst.MaxRestarts = 0
		}
		inst.Restarts = new(int32)
		// Which signal should be used to stop (i.e. exit gracefully) the program
		inst.StopSignal = inst.ParseSignal(cfg, STOPSIGNALMSG)
		// How long to wait after a graceful stop before killing the program
		// inst.ParseInt(cfg, "StopTimeout", 1, STOPTIMEOUTMSG)
		if timeout, err := strconv.Atoi(cfg.StopTimeout); err == nil {
			inst.StopTimeout = timeout
		} else {
			inst.StopTimeout = 5
		}
		// Options to discard the program’s stdout/stderr or to redirect them to files
		inst.ParseRedirections(cfg)
		// Environment variables to set before launching the program
		inst.EnvVars = strings.Fields(cfg.EnvVars)
		// A working directory to set before launching the program
		inst.WorkingDir = cfg.WorkingDir
		// An umask to set before launching the program
		inst.ParseInt(cfg, "Umask", defaultUmask, UMASKMSG)
		// Add conditional var to struct
		inst.condition = sync.NewCond(&inst.mutex)
		inst.finishedCh = make(chan struct{})
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
