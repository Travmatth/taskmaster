package main

import (
	"io/ioutil"
	"strings"
	"syscall"

	"gopkg.in/yaml.v2"
)

const (
	EXPECTEDEXITMSG = "Error: expectedExit must be signal name found in `man signal`"
	STOPSIGNALMSG   = "Error: stopSignal must be signal name found in `man signal`"
	INSTANCESMSG    = "Error: invalid instances value: %s\n"
	STARTCHECKUPMSG = "Error: invalid startCheckup value: %s\n"
	MAXRESTARTSMSG  = "Error: invalid maxrestarts value: %s\n"
	KILLTIMEOUTMSG  = "Error: invalid killtimeout value: %s\n"
	UMASKMSG        = "Error: invalid umask value: %s\n"
)

// ProcConfig represents the config struct loaded from yaml
type ProcConfig struct {
	ID            string
	Command       string
	Instances     string
	AtLaunch      string
	RestartPolicy string
	ExpectedExit  string
	StartCheckup  string
	MaxRestarts   string
	StopSignal    string
	KillTimeout   string
	EnvVars       string
	WorkingDir    string
	Umask         string
	pid           string
	status        string
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

// SetDefaults translate ProcConfig array into Proc array, verifying inputs and setting defaults
func SetDefaults(configProcs []ProcConfig) []Proc {
	ids := make(map[int]bool)
	procs := []Proc{}
	defaultUmask := syscall.Umask(0)

	syscall.Umask(defaultUmask)
	for _, cfg := range configProcs {
		var proc Proc
		// an id to uniquely identify the process
		proc.ParseID(&cfg, ids)
		// The command to use to launch the program
		proc.Args = strings.Fields(cfg.Command)
		// The number of processes to start and keep running
		proc.ParseInt(&cfg, "Instances", 1, INSTANCESMSG)
		// Whether to start this program at launch or not
		proc.ParseAtLaunch(&cfg)
		// Whether the program should be restarted always, never, or on unexpected exits only
		proc.ParseRestartPolicy(&cfg)
		// Which return codes represent an "expected" exit status
		proc.ParseInt(&cfg, "ExpectedExit", 0, EXPECTEDEXITMSG)
		// How long the program should be running after it’s started for it to be considered "successfully started"
		proc.ParseInt(&cfg, "StartCheckup", 1, STARTCHECKUPMSG)
		// How many times a restart should be attempted before aborting
		proc.ParseInt(&cfg, "MaxRestarts", 0, MAXRESTARTSMSG)
		// Which signal should be used to stop (i.e. exit gracefully) the program
		proc.StopSignal = proc.ParseSignal(&cfg, STOPSIGNALMSG)
		// How long to wait after a graceful stop before killing the program
		proc.ParseInt(&cfg, "KillTimeout", 1, KILLTIMEOUTMSG)
		// Options to discard the program’s stdout/stderr or to redirect them to files
		proc.ParseRedirections(&cfg)
		// Environment variables to set before launching the program
		proc.EnvVars = strings.Fields(cfg.EnvVars)
		// A working directory to set before launching the program
		proc.WorkingDir = cfg.WorkingDir
		// An umask to set before launching the program
		proc.ParseInt(&cfg, "Umask", defaultUmask, UMASKMSG)
		// Add proc to array
		procs = append(procs, proc)
	}
	return procs
}

// LoadConfig opens config file and parses yaml syntax into array of ProcConfig structs
func LoadConfig(yamlFile string) []ProcConfig {
	configs := []ProcConfig{}
	buf, err := ioutil.ReadFile(yamlFile)
	Check(err)
	err = yaml.Unmarshal(buf, &configs)
	Check(err)
	return configs
}
