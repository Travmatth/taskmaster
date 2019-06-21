package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Proc structs contain basic information of launched processes
type Proc struct {
	Command       string
	Instances     int
	AtLaunch      bool
	RestartPolicy string
	ExpectedExit  string
	StartCheckup  string
	MaxRestarts   string
	StopSignal    string
	KillTimeout   int
	Redirections  struct {
		Stdout string
		Stderr string
	}
	EnvVars    string
	WorkingDir string
	Umask      string
	start      time.Time
	Process    os.Process
}

// Check absence of errors
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func loadYamlFile(yamlFile string) []byte {
	buf, err := ioutil.ReadFile(yamlFile)
	Check(err)
	return buf
}

func loadConfig(yamlFile string) []Proc {
	config := []Proc{}
	file := loadYamlFile(yamlFile)
	err := yaml.Unmarshal(file, &config)
	Check(err)
	return config
}

func main() {
	if l := len(os.Args); l < 2 {
		fmt.Print("Error: Please specify configuration file(s)\n")
	} else {
		var procs []Proc
		for i := 1; i < l; i++ {
			yamlFile := os.Args[i]
			procs = append(procs, loadConfig(yamlFile)...)
		}
		fmt.Print(procs)
	}
}
