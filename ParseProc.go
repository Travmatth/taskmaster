package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"gopkg.in/oleiade/reflections.v1"
)

/*
ParseID parses the config struct to set the Job ID
or exit with error if incorrectly set
*/
func (p *Job) ParseID(c *JobConfig, ids map[int]bool) {
	if c.ID == "" {
		fmt.Println("Error: ID must be specified")
	} else if val, err := strconv.Atoi(c.ID); err != nil {
		fmt.Println("Error: ID must be integer")
	} else if _, ok := ids[val]; ok {
		fmt.Println("Error: ID must be unique")
	} else {
		ids[val] = true
		p.ID = val
		return
	}
	os.Exit(1)
}

/*
ParseInt uses https://golang.org/pkg/reflect/ to dynamically set struct member to integer
*/
func (p *Job) ParseInt(c *JobConfig, member string, defaultVal int, message string) {
	cfgVal, _ := reflections.GetField(c, member)

	if cfgVal == "" {
		p.Umask = defaultVal
	} else if val, err := strconv.Atoi(cfgVal.(string)); err != nil {
		fmt.Printf(message, cfgVal)
		os.Exit(1)
	} else {
		reflections.SetField(p, member, val)
	}
}

/*
ParseatLaunch parses the config struct to set the at launch policy
*/
func (j *Job) ParseAtLaunch(c *JobConfig) {
	if c.AtLaunch == "" || strings.ToLower(c.AtLaunch) == "true" {
		j.AtLaunch = true
	} else {
		j.AtLaunch = false
	}
}

/*
ParserestartPolicy parses the config struct to set the restart policy
or exit with error if incorrectly set
*/
func (p *Job) ParserestartPolicy(c *JobConfig) {
	policy := strings.ToLower(c.RestartPolicy)
	switch policy {
	case "always":
		p.restartPolicy = RESTARTALWAYS
	case "never":
		p.restartPolicy = RESTARTNEVER
	case "unexpected":
		p.restartPolicy = RESTARTUNEXPECTED
	default:
		fmt.Println("Error: Resart Policy must be one of: always | never | unexpected, recieved: ", c)
		os.Exit(1)
	}
}

/*
ParseSignal parses the config struct to set the signal of the given Job member
or exit with error if incorrectly set
*/
func (p *Job) ParseSignal(c *JobConfig, message string) syscall.Signal {
	if sig, ok := Signals[strings.ToUpper(c.StopSignal)]; ok {
		return sig
	}
	os.Exit(1)
	return syscall.Signal(0)
}

//OpenRedir opens the given file for use in Jobess's redirections
func (p *Job) OpenRedir(val string, flag int) *os.File {
	if val != "" {
		f, err := os.OpenFile(val, flag, 0666)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return f
	}
	return nil
}

/*
ParseRedirections parses the config struct to open the given filename and set *File member of
Job struct, raises runtime error if incorrectly set
*/
func (p *Job) ParseRedirections(c *JobConfig) {
	base := os.O_CREATE
	p.Redirections = []*os.File{
		p.OpenRedir(c.Redirections.Stdin, base|os.O_RDONLY),
		p.OpenRedir(c.Redirections.Stdout, base|os.O_WRONLY|os.O_TRUNC),
		p.OpenRedir(c.Redirections.Stderr, base|os.O_WRONLY|os.O_TRUNC),
	}
}
