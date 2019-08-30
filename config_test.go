package main

import (
	"syscall"
	"testing"
)

func TestConfigOpenRedirIgnoresEmptyFile(t *testing.T) {
	if f, err := OpenRedir("", 0); err != nil {
		t.Errorf("OpenRedir should not return an error on empty file")
	} else if f != nil {
		t.Errorf("OpenRedir should not return file on empty file")
	}
	Buf.Reset()
}

func TestConfigOpenRedirErrorsOnBadFile(t *testing.T) {
	if f, err := OpenRedir("foo", 0); err == nil {
		t.Errorf("OpenRedir should return an error on bad file")
	} else if f != nil {
		t.Errorf("OpenRedir should not return file on bad file")
	}
	Buf.Reset()
}

func TestConfigOpenRedirOpensFile(t *testing.T) {
	if f, err := OpenRedir("procfiles/basic.yaml", 0); err != nil {
		t.Errorf("OpenRedir should open valid file without error")
	} else if f == nil {
		t.Errorf("OpenRedir should open valid file")
	} else {
		f.Close()
	}
	Buf.Reset()
}

func TestConfigParseIntSetsDefaultVal(t *testing.T) {
	var c JobConfig
	var i Instance

	if err := ParseInt(c, &i, "Umask", 2, ""); err != nil {
		t.Errorf("ParseInt should not return an error when setting default values")
	} else if i.Umask != 2 {
		t.Errorf("ParseInt should set default values when config empty")
	}
	Buf.Reset()
}

func TestConfigParseIntDetectsInvalidInput(t *testing.T) {
	var c JobConfig
	var i Instance

	c.Umask = "foo"
	if err := ParseInt(c, &i, "Umask", 2, ""); err == nil {
		t.Errorf("ParseInt should return an error when setting invalid values")
	}
	Buf.Reset()
}

func TestConfigParseIntSetsValue(t *testing.T) {
	var c JobConfig
	var i Instance

	c.Umask = "3"
	if err := ParseInt(c, &i, "Umask", 2, ""); err != nil {
		t.Errorf("ParseInt should not return an error when setting valid values")
	} else if i.Umask != 3 {
		t.Errorf("ParseInt should set valid values")
	}
	Buf.Reset()
}

func TestConfigConfigureInstance(t *testing.T) {
	var i Instance
	c := JobConfig{
		ID:            "0",
		Command:       "foo",
		Instances:     "",
		AtLaunch:      "",
		RestartPolicy: "",
		ExpectedExit:  "",
		StartCheckup:  "",
		MaxRestarts:   "",
		StopSignal:    "",
		StopTimeout:   "",
		EnvVars:       "",
		WorkingDir:    "",
		Umask:         "",
		Redirections: Redirections{
			Stdin:  "",
			Stdout: "",
			Stderr: "",
		},
	}

	if err := ConfigureInstance(c, &i, 0); err != nil {
		t.Errorf("ConfigureInstance should parse a valid configuration struct %s", err.Error())
	}
	if len(i.args) != 1 || i.args[0] != "foo" {
		t.Errorf("ConfigureInstance doesnt correctly set default args")
	} else if i.restartPolicy != RESTARTNEVER {
		t.Errorf("ConfigureInstance doesnt correctly set default restartPolicy")
	} else if i.ExpectedExit != 0 {
		t.Errorf("ConfigureInstance doesnt correctly set default ExpectedExit")
	} else if i.StartCheckup != 0 {
		t.Errorf("ConfigureInstance doesnt correctly set default StartCheckup")
	} else if *i.Restarts != 0 {
		t.Errorf("ConfigureInstance doesnt correctly set default Restarts")
	} else if i.MaxRestarts != 0 {
		t.Errorf("ConfigureInstance doesnt correctly set default MaxRestarts")
	} else if i.StopSignal != syscall.Signal(0) {
		t.Errorf("ConfigureInstance doesnt correctly set default StopSignal")
	} else if i.StopTimeout != 1 {
		t.Errorf("ConfigureInstance doesnt correctly set default StopTimeout")
	} else if len(i.EnvVars) != 0 {
		t.Errorf("ConfigureInstance doesnt correctly set default EnvVars")
	} else if i.WorkingDir != "" {
		t.Errorf("ConfigureInstance doesnt correctly set default WorkingDir")
	} else if i.Umask != 0 {
		t.Errorf("ConfigureInstance doesnt correctly set default Umask")
	} else if len(i.Redirections) != 3 || (i.Redirections[0] != nil &&
		i.Redirections[1] != nil &&
		i.Redirections[2] != nil) {
		t.Errorf("ConfigureInstance doesnt correctly set default Redirections")
	}
	Buf.Reset()
}

func TestConfigConfigureJob(t *testing.T) {
	var j Job
	ids := make(map[int]bool)
	c := JobConfig{
		ID:            "0",
		Command:       "foo",
		Instances:     "",
		AtLaunch:      "",
		RestartPolicy: "",
		ExpectedExit:  "",
		StartCheckup:  "",
		MaxRestarts:   "",
		StopSignal:    "",
		StopTimeout:   "",
		EnvVars:       "",
		WorkingDir:    "",
		Umask:         "",
		Redirections: Redirections{
			Stdin:  "",
			Stdout: "",
			Stderr: "",
		},
	}

	if err := ConfigureJob(c, &j, ids); err != nil {
		t.Errorf("ConfigureJob should parse a valid configuration struct %s", err.Error())
	} else if j.ID != 0 {
		t.Errorf("ConfigureJob doesnt correctly set ID")
	} else if len(j.Instances) != 1 {
		t.Errorf("ConfigureJob doesnt correctly set Instances")
	} else if j.pool != 1 {
		t.Errorf("ConfigureJob doesnt correctly set pool")
	} else if *j.cfg != c {
		t.Errorf("ConfigureJob doesnt correctly set cfg")
	} else if j.AtLaunch != true {
		t.Errorf("ConfigureJob doesnt correctly set AtLaunch")
	}
	Buf.Reset()
}

func TestConfigConfigureJobShouldErrorOnRepeatID(t *testing.T) {
	var j Job
	ids := make(map[int]bool)
	ids[0] = true
	c := JobConfig{
		ID:            "0",
		Command:       "foo",
		Instances:     "",
		AtLaunch:      "",
		RestartPolicy: "",
		ExpectedExit:  "",
		StartCheckup:  "",
		MaxRestarts:   "",
		StopSignal:    "",
		StopTimeout:   "",
		EnvVars:       "",
		WorkingDir:    "",
		Umask:         "",
		Redirections: Redirections{
			Stdin:  "",
			Stdout: "",
			Stderr: "",
		},
	}

	if err := ConfigureJob(c, &j, ids); err == nil {
		t.Errorf("ConfigureJob should error on nonunique ID")
	}
	Buf.Reset()
}

func TestConfigSetDefaults(t *testing.T) {
	Buf.Reset()
}

func TestConfigLoadFile(t *testing.T) {
	Buf.Reset()
}

func TestConfigLoadJobs(t *testing.T) {
	Buf.Reset()
}

func TestConfigLoadJobsFromFile(t *testing.T) {
	Buf.Reset()
}
