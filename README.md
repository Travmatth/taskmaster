# TaskMaster: A MacOS process manager
`Usage: ./taskmaster <Config_File> <Log_File> [Log_level]`
```
Usage: ./taskmaster <Config_File> <Log_File> [Log_Level]
        Config_File: Procfile you wish to run
        Log_File: Log file you wish to use
        Log_Level:  0 CRITICAL, 1 ERROR, 2 WARNING, 3 NOTICE, 4 INFO, 5 DEBUG
```

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
exit:       stop all jobs and exit taskmaster
```