package parse

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"

	CFG "github.com/Travmatth/taskmaster/config"
	INST "github.com/Travmatth/taskmaster/instance"
	JOB "github.com/Travmatth/taskmaster/job"
	. "github.com/Travmatth/taskmaster/log"
	SIG "github.com/Travmatth/taskmaster/signals"
	"gopkg.in/oleiade/reflections.v1"
	"gopkg.in/yaml.v2"
)

// Exit messages for the various parsing errors
const (
	RESTARTERRMSG   = "Error: Restart Policy for %v must be one of: always | never | unexpected, recieved: \"%s\""
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

//ConfigureInstance parse configuration file to set Instance struct properties
func ConfigureInstance(c CFG.JobConfig,
	instance *INST.Instance, umask int) error {
	// The command to use to launch the program
	instance.Args = strings.Fields(c.Command)
	// Whether the program should restart always, never, or unexpected exits
	switch strings.ToLower(c.RestartPolicy) {
	case "always":
		instance.RestartPolicy = INST.RESTARTALWAYS
	case "never":
		instance.RestartPolicy = INST.RESTARTNEVER
	case "unexpected":
		instance.RestartPolicy = INST.RESTARTUNEXPECTED
	case "":
		Log.Info(c, "does not specify a restart policy, NEVER set as default")
		instance.RestartPolicy = INST.RESTARTNEVER
	default:
		return fmt.Errorf(RESTARTERRMSG, c, strings.ToLower(c.RestartPolicy))
	}
	// Which return codes represent an "expected" exit Status
	ParseInt(c, instance, "ExpectedExit", EXPECTEDEXITMSG, 0)
	// How long the program should be running after it’s started for
	// it to beconsidered "successfully started"
	ParseInt(c, instance, "StartCheckup", STARTCHECKUPMSG, 0)
	// How many times a restart should be attempted before aborting
	if c.MaxRestarts == "" {
		instance.MaxRestarts = 0
	} else if val, err := strconv.Atoi(c.MaxRestarts); err != nil {
		return fmt.Errorf("%v Error: invalid maxrestarts value", c)
	} else {
		instance.MaxRestarts = int32(val)
	}
	instance.Restarts = new(int32)
	// signal used to stop (instance.e. exit gracefully) the program
	if c.StopSignal == "" {
		instance.StopSignal = syscall.Signal(0)
	} else if sig, ok := SIG.Signals[strings.ToUpper(c.StopSignal)]; ok {
		instance.StopSignal = sig
	} else {
		return fmt.Errorf("Configuration error: invalid stop signal for %v", c)
	}
	// How long to wait after a graceful stop before killing the program
	ParseInt(c, instance, "StopTimeout", STOPTIMEOUTMSG, 1)
	// Options to discard stdout/stderr or to redirect them to files
	in := c.Redirections.Stdin
	out := c.Redirections.Stdout
	serr := c.Redirections.Stderr
	if stdin, err := OpenRedir(in, stdinFlags); err != nil {
		return err
	} else if stdout, err := OpenRedir(out, stdoutFlags); err != nil {
		return err
	} else if stderr, err := OpenRedir(serr, stdoutFlags); err != nil {
		return err
	} else {
		instance.Redirections = []*os.File{stdin, stdout, stderr}
	}
	// Environment variables to set before launching the program
	instance.EnvVars = strings.Fields(c.EnvVars)
	// A working directory to set before launching the program
	instance.WorkingDir = c.WorkingDir
	// An umask to set before launching the program
	ParseInt(c, instance, "Umask", UMASKMSG, umask)
	// Add conditional var to struct
	instance.Condition = sync.NewCond(&instance.Mutex)
	instance.FinishedCh = make(chan struct{})
	return nil
}

//ConfigureJob parse configuration file to set Job struct properties
func ConfigureJob(c CFG.JobConfig, job *JOB.Job, ids map[int]bool) error {
	var message string
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
		job.Pool = 1
	} else if pool, err := strconv.Atoi(c.Instances); err != nil {
		message = "Configuration error: invalid instances value for Job ID: %d"
		return fmt.Errorf(message, job.ID)
	} else {
		job.Pool = pool
	}
	// Add config to job struct
	job.Cfg = &c
	job.Instances = make([]*INST.Instance, job.Pool)
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
		message = "%v configuration error: invalid value for atLaunch"
		return fmt.Errorf(message, c)
	}
	return nil
}

/*
 * SetDefaults translate []JobConfig -> []Job, verifying inputs/setting defaults
 */
func SetDefaults(configJobs []CFG.JobConfig) ([]*JOB.Job, error) {
	ids := make(map[int]bool)
	jobs := []*JOB.Job{}
	umask := GetDefaultUmask()
	for _, c := range configJobs {
		var job JOB.Job
		if err := ConfigureJob(c, &job, ids); err != nil {
			return nil, err
		}
		for i := 0; i < job.Pool; i++ {
			var instance INST.Instance
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

/*
 * OpenRedir opens the given file for use in Jobess's redirections
 */
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
 * LoadJobsFromFile loads a configuration file and configures given jobs
 */
func LoadJobsFromFile(file string) ([]*JOB.Job, error) {
	if buf, fileErr := LoadFile(file); fileErr != nil {
		return nil, fileErr
	} else if jobConfigs, configErr := LoadJobs(buf); configErr != nil {
		return nil, configErr
	} else {
		return SetDefaults(jobConfigs)
	}
}

/*
 * LoadFile Reads given Procfile into buffer
 */
func LoadFile(file string) ([]byte, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

/*
 * LoadJobs opens config and parses yaml syntax
 */
func LoadJobs(buf []byte) ([]CFG.JobConfig, error) {
	configs := []CFG.JobConfig{}
	if err := yaml.Unmarshal(buf, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

/*
 * GetDefaultUmask obtains current file permissions
 */
func GetDefaultUmask() int {
	defaultUmask := syscall.Umask(0)
	syscall.Umask(defaultUmask)
	return defaultUmask
}

/*
 * ParseInt uses [1] to dynamically set struct member to integer
 * [1]https://golang.org/pkg/reflect/
 */
func ParseInt(c CFG.JobConfig,
	instance *INST.Instance, member, message string, defaultVal int) error {
	str, _ := reflections.GetField(c, member)
	if str == "" {
		reflections.SetField(instance, member, defaultVal)
	} else if num, err := strconv.Atoi(str.(string)); err != nil {
		return fmt.Errorf(message, c)
	} else {
		reflections.SetField(instance, member, num)
	}
	return nil
}
