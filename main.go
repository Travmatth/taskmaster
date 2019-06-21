package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

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
	var supervisor Supervisor

	if l := len(os.Args); l != 2 {
		fmt.Print("Error: Please specify a configuration file\n")
	} else {
		yamlFile := os.Args[1]
		supervisor.procs = loadConfig(yamlFile)
		supervisor.Start()
	}
}
