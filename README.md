# TaskMaster: A process manager

`Usage: ./taskmaster <Config_File> <Log_File> [Log_level]`

Taskmaster accepts a config file containing a list of processes to start, along with the options managing their execution and termination. Provides a simple UI to manage processes.

# Configuration:

```
- id: an ID to identify the process, must be unique
  command: [int] command & options to be executed
  instances: [int] number of instances to launch
  atLaunch: [bool=true] [default=true] whether to launch at startup
  restartPolicy: [always|unexpected|never] whether to restart instances always|never|unexpected exit
  expectedExit: [int] the expected exit code
  startCheckup: [int] time in seconds to wait before checking if the process started successfully
  maxRestarts: [int] the maximum number of times to attempt restart if failed
  stopSignal: [string] signal to be sent to process to kill (name in `man signal`)
  stopTimeout: [int] time in seconds to wait after sending stop signal before manually killing the process
  redirections:
    stdin: [string] file to redirect stdin
    stdout: [string] file to redirect stdout
    stderr: [string] file to redirect stderr
  envVars: [string] "name=val name2=val2" variables to provide to the process environment
  workingDir: [string] a path to set as the current working directory
  umask: [int] umask to set the process permissions
```

# UI Commands

```
ps:         List current jobs being managed
logs:       display jobs logs
clear:      clear the screen
start [id]: start given job
stop [id]:  stop given job
startAll:   start all jobs
stopAll:    stop all jobs
reload:     reload the configuration file
exit:       stop all jobs and exit taskmaste
```

# Requirements:

- [x] See the Status of all the programs described in the config file ("Status" command)
- [x] Start / stop / restart programs
- [x] Reload the configuration file without stopping the main program
- [x] Stop the main program
The configuration file must allow the user to specify the following, for each program that is to b be supervised:
- [x] The command to use to launch the program
- [x] The number of processes to start and keep running
- [x] Whether to start this program at launch or not
- [x] Whether the program should be restarted always, never, or on unexpected exits only
- [x] Which return codes represent an "expected" exit Status
- [x] How long the program should be running after it’s started for it to be considered "successfully started"
- [x] How many times a restart should be attempted before aborting
- [x] Which signal should be used to stop (i.e. exit gracefully) the program
- [x] How long to wait after a graceful stop before killing the program
- [x] Options to discard the program’s stdout/stderr or to redirect them to files
- [x] Environment variables to set before launching the program
- [x] A working directory to set before launching the program
- [x] An umask to set before launching the program