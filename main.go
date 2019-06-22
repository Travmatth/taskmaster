package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"gopkg.in/yaml.v2"
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

// Check absence of errors
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// parseInt uses https://golang.org/pkg/reflect/ to dynamically set struct member
func parseInt(cfg *ProcConfig, proc *Proc, defaultVal int, member string, message string) {
	cfgReflect := reflect.ValueOf(cfg)
	procReflect := reflect.ValueOf(proc)
	cfgVal := reflect.Indirect(cfgReflect).FieldByName(member).String()

	if cfgVal == "" {
		proc.Umask = defaultVal
	} else if val, err := strconv.Atoi(cfgVal); err != nil {
		fmt.Printf(message, cfgVal)
		os.Exit(1)
	} else {
		procReflect.Elem().FieldByName(member).SetInt(int64(val))

	}
}
func setDefaults(configProcs []ProcConfig) []Proc {
	procs := []Proc{}
	defaultUmask := syscall.Umask(0)

	syscall.Umask(defaultUmask)
	for _, cfg := range configProcs {
		var proc Proc
		proc.Args = strings.Fields(cfg.Command)
		// check id, err exit if not set
		parseInt(&cfg, &proc, defaultUmask, "Umask", "Error: invalid umask value: %s\n")
	}
	return procs
}

func loadConfig(yamlFile string) []Proc {
	configs := []ProcConfig{}
	buf, err := ioutil.ReadFile(yamlFile)
	Check(err)
	err = yaml.Unmarshal(buf, &configs)
	Check(err)
	return setDefaults(configs)
}

func main() {
	var supervisor Supervisor

	if l := len(os.Args); l != 2 {
		fmt.Print("Error: Please specify a configuration file\n")
	} else {
		yamlFile := os.Args[1]
		newProcs := loadConfig(yamlFile)
		supervisor.StartAll(newProcs)
	}
}
