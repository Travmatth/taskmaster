package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"gopkg.in/oleiade/reflections.v1"
	"gopkg.in/yaml.v2"
)

// Exit messages for the various parsing errors
const (
	EXPECTEDEXITMSG = "Error: expectedExit must be signal name found in `man signal`"
	STOPSIGNALMSG   = "Error: stopSignal must be signal name found in `man signal`"
	STARTCHECKUPMSG = "Error: invalid startCheckup value: %s\n"
	STOPTIMEOUTMSG  = "Error: invalid StopTimeout value: %s\n"
	UMASKMSG        = "Error: invalid umask value: %s\n"
)

// Flags used in OpenRedir
const (
	stdinFlags  = os.O_CREATE | os.O_RDONLY
	stdoutFlags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	stderrFlags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
)

// Redirections store the redirection file names
type Redirections struct {
	Stdin  string `json:"Stdin"`
	Stdout string `json:"Stdout"`
	Stderr string `json:"Stderr"`
}

// JobConfig represents the config struct loaded from yaml
type JobConfig struct {
	ID            string `json:"ID" yaml:"id"`
	Command       string `json:"Command" yaml:"command"`
	Instances     string `json:"Instances" yaml:"instances"`
	AtLaunch      string `json:"AtLaunch" yaml:"atLaunch"`
	RestartPolicy string `json:"RestartPolicy" yaml:"restartPolicy"`
	ExpectedExit  string `json:"ExpectedExit" yaml:"expectedExit"`
	StartCheckup  string `json:"StartCheckup" yaml:"startCheckup"`
	MaxRestarts   string `json:"MaxRestarts" yaml:"maxRestarts"`
	StopSignal    string `json:"StopSignal" yaml:"stopSignal"`
	StopTimeout   string `json:"StopTimeout" yaml:"stopTimeout"`
	EnvVars       string `json:"EnvVars" yaml:"envVars"`
	WorkingDir    string `json:"WorkingDir" yaml:"workingDir"`
	Umask         string `json:"Umask" yaml:"umask"`
	Redirections
}

//GetDefaultUmask obtains current file permissions
func GetDefaultUmask() int {
	defaultUmask := syscall.Umask(0)
	syscall.Umask(defaultUmask)
	return defaultUmask
}

//OpenRedir opens the given file for use in Jobess's redirections
func OpenRedir(val string, flag int) (*os.File, error) {
	if val != "" {
		f, err := os.OpenFile(val, flag, 0666)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
	return nil, nil
}

/*
ParseInt uses https://golang.org/pkg/reflect/ to dynamically set struct member to integer
*/
func ParseInt(c JobConfig, i *Instance, member string, defaultVal int, message string) error {
	cfgVal, _ := reflections.GetField(c, member)

	if cfgVal == "" {
		reflections.SetField(i, member, defaultVal)
	} else if val, err := strconv.Atoi(cfgVal.(string)); err != nil {
		return fmt.Errorf(message, c)
	} else {
		reflections.SetField(i, member, val)
	}
	return nil
}

//ConfigureInstance parse configuration file to set Instance struct properties
func ConfigureInstance(c JobConfig, i *Instance, umask int) error {
	// The command to use to launch the program
	i.args = strings.Fields(c.Command)
	// Whether the program should be restarted always, never, or on unexpected exits only
	switch strings.ToLower(c.RestartPolicy) {
	case "always":
		i.restartPolicy = RESTARTALWAYS
	case "never":
		i.restartPolicy = RESTARTNEVER
	case "unexpected":
		i.restartPolicy = RESTARTUNEXPECTED
	case "":
		Log.Info(c, "does not specify a restart policy, NEVER set as default")
		i.restartPolicy = RESTARTNEVER
	default:
		return fmt.Errorf("Error: Resart Policy for %v must be one of: always | never | unexpected, recieved: \"%s\"", c, strings.ToLower(c.RestartPolicy))
	}
	// Which return codes represent an "expected" exit Status
	ParseInt(c, i, "ExpectedExit", 0, EXPECTEDEXITMSG)
	// How long the program should be running after it’s started for it to be considered "successfully started"
	ParseInt(c, i, "StartCheckup", 0, STARTCHECKUPMSG)
	// How many times a restart should be attempted before aborting
	if c.MaxRestarts == "" {
		i.MaxRestarts = 0
	} else if val, err := strconv.Atoi(c.MaxRestarts); err != nil {
		return fmt.Errorf("%v Error: invalid maxrestarts value\n", c)
	} else {
		i.MaxRestarts = int32(val)
	}
	i.Restarts = new(int32)
	// Which signal should be used to stop (i.e. exit gracefully) the program
	if c.StopSignal == "" {
		i.StopSignal = syscall.Signal(0)
	} else if sig, ok := Signals[strings.ToUpper(c.StopSignal)]; ok {
		i.StopSignal = sig
	} else {
		return fmt.Errorf("Configuration error: invalid stop signal for %v", c)
	}
	// How long to wait after a graceful stop before killing the program
	ParseInt(c, i, "StopTimeout", 1, STOPTIMEOUTMSG)
	// Options to discard the program’s stdout/stderr or to redirect them to files
	if stdin, err := OpenRedir(c.Redirections.Stdin, stdinFlags); err != nil {
		return err
	} else if stdout, err := OpenRedir(c.Redirections.Stdout, stdoutFlags); err != nil {
		return err
	} else if stderr, err := OpenRedir(c.Redirections.Stderr, stdoutFlags); err != nil {
		return err
	} else {
		i.Redirections = []*os.File{stdin, stdout, stderr}
	}
	// Environment variables to set before launching the program
	i.EnvVars = strings.Fields(c.EnvVars)
	// A working directory to set before launching the program
	i.WorkingDir = c.WorkingDir
	// An umask to set before launching the program
	ParseInt(c, i, "Umask", umask, UMASKMSG)
	// Add conditional var to struct
	i.condition = sync.NewCond(&i.mutex)
	i.finishedCh = make(chan struct{})
	return nil
}

//ConfigureJob parse configuration file to set Job struct properties
func ConfigureJob(c JobConfig, job *Job, ids map[int]bool) error {
	// an id to uniquely identify the Jobs
	if c.ID == "" {
		return fmt.Errorf("Error: ID must be specified")
	} else if val, err := strconv.Atoi(c.ID); err != nil {
		return fmt.Errorf("Error: ID must be integer")
	} else if _, ok := ids[val]; ok {
		return fmt.Errorf("Error: ID must be unique")
	} else {
		ids[val] = true
		job.ID = val
	}
	// The number of Instances to start and keep running
	if c.Instances == "" {
		job.pool = 1
	} else if pool, err := strconv.Atoi(c.Instances); err != nil {
		return fmt.Errorf("Configuration error: invalid instances value for Job ID: %d", job.ID)
	} else {
		job.pool = pool
	}
	// Add config to job struct
	job.cfg = &c
	job.Instances = make([]*Instance, job.pool)
	// Whether to start this program at launch or not
	switch strings.ToLower(c.AtLaunch) {
	case "true":
		job.AtLaunch = true
	case "false":
		job.AtLaunch = false
	case "":
		Log.Info(job, "AtLaunch empty, defaulting to true")
		job.AtLaunch = true
	default:
		return fmt.Errorf("%v configuration error: invalid value for atLaunch", c)
	}
	return nil
}

// SetDefaults translate JobConfig array into Job array, verifying inputs and setting defaults
func SetDefaults(configJobs []JobConfig) ([]*Job, error) {
	ids := make(map[int]bool)
	jobs := []*Job{}
	umask := GetDefaultUmask()
	for _, c := range configJobs {
		var job Job
		if err := ConfigureJob(c, &job, ids); err != nil {
			return nil, err
		}
		for i := 0; i < job.pool; i++ {
			var instance Instance

			instance.JobID = job.ID
			instance.InstanceID = i
			job.Instances[i] = &instance
			if err := ConfigureInstance(c, &instance, umask); err != nil {
				return nil, err
			}
		}
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

//LoadFile Reads given Procfile into buffer
func LoadFile(file string) ([]byte, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// LoadJobs opens config file and parses yaml syntax into array of JobConfig structs
func LoadJobs(buf []byte) ([]JobConfig, error) {
	configs := []JobConfig{}
	if err := yaml.Unmarshal(buf, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// LoadJobsFromFile loads a configuration file and configures given jobs
func LoadJobsFromFile(file string) ([]*Job, error) {
	if buf, fileErr := LoadFile(file); fileErr != nil {
		return nil, fileErr
	} else if jobConfigs, configErr := LoadJobs(buf); configErr != nil {
		return nil, configErr
	} else {
		return SetDefaults(jobConfigs)
	}
}

func (c JobConfig) Same(cfg *JobConfig) bool {
	if c.ID != cfg.ID ||
		c.Command != cfg.Command ||
		c.Instances != cfg.Instances ||
		c.AtLaunch != cfg.AtLaunch ||
		c.RestartPolicy != cfg.RestartPolicy ||
		c.ExpectedExit != cfg.ExpectedExit ||
		c.StartCheckup != cfg.StartCheckup ||
		c.MaxRestarts != cfg.MaxRestarts ||
		c.StopSignal != cfg.StopSignal ||
		c.StopTimeout != cfg.StopTimeout ||
		c.EnvVars != cfg.EnvVars ||
		c.WorkingDir != cfg.WorkingDir ||
		c.Umask != cfg.Umask ||
		c.Redirections.Stdin != cfg.Redirections.Stdin ||
		c.Redirections.Stdout != cfg.Redirections.Stdout ||
		c.Redirections.Stderr != cfg.Redirections.Stderr {
		return false
	}
	return true
}

func (c JobConfig) String() string {
	return fmt.Sprintf("JobConfig %s", c.ID)
}
