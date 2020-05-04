package config

import (
	"fmt"
)

/*
 * Redirections store the redirection file names
 */
type Redirections struct {
	Stdin  string `json:"Stdin"`
	Stdout string `json:"Stdout"`
	Stderr string `json:"Stderr"`
}

/*
 * JobConfig represents the config struct loaded from yaml
 */
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

/*
 * Same compares two configuration files for deep equality
 */
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

/*
 * String is the printed representation of the struct
 */
func (c JobConfig) String() string {
	return fmt.Sprintf("JobConfig %s", c.ID)
}
