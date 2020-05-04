package ui

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	INST "github.com/Travmatth/taskmaster/instance"
	JOB "github.com/Travmatth/taskmaster/job"
	SIG "github.com/Travmatth/taskmaster/signals"
	S "github.com/Travmatth/taskmaster/supervisor"
)

type Frontend struct {
	supervisor *S.Supervisor
	scanner    *bufio.Scanner
}

func NewFrontend(supervisor *S.Supervisor) (f *Frontend) {
	f = &Frontend{
		supervisor: supervisor,
		scanner:    bufio.NewScanner(os.Stdin),
	}
	return
}

func (f *Frontend) StartUI() {
	fmt.Print("> ")
UILoop:
	for f.scanner.Scan() {
		input := strings.ToLower(f.scanner.Text())
		end := f.DecideCommand(input)
		if end == true {
			break UILoop
		}
		fmt.Print("> ")
	}
	if err := f.scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintln(os.Stderr, err)
		f.supervisor.SigCh <- SIG.Signals["SIGTERM"]
	}
}

func (f *Frontend) DecideCommand(input string) bool {
	switch {
	case input == "exit":
		fmt.Println("Exiting TaskMaster")
		f.supervisor.SigCh <- SIG.Signals["SIGTERM"]
		return true
	case input == "reload":
		fmt.Println("Reloading TaskMaster config")
		f.supervisor.SigCh <- SIG.Signals["SIGHUP"]
	case input == "logs":
		f.PrintLogs()
	case input == "clear":
		command := exec.Command("clear")
		command.Stdout = os.Stdout
		command.Run()
	case input == "startall" || input == "start all":
		fmt.Println("Starting all jobs")
		f.supervisor.StartAllJobs(false)
	case input == "stopall" || input == "stop all":
		fmt.Println("Stopping all jobs")
		f.supervisor.StopAllJobs(false)
	case strings.HasPrefix(input, "start"):
		f.WithId(input, func(id int) {
			fmt.Println("Starting", id)
			f.supervisor.StartJob(id, false)
		})
	case strings.HasPrefix(input, "stop"):
		f.WithId(input, func(id int) {
			fmt.Println("Stopping", id)
			f.supervisor.StopJob(id)
		})
	case strings.HasPrefix(input, "ps"):
		format := "%-12s%-12s%-12s%-12s\n"
		fmt.Printf(format, "ID", "Instance", "PID", "Status")
		fmt.Print(f.FormatJobs())
	case strings.HasPrefix(input, "help"):
		f.PrintHelp()
	default:
		return false
	}
	return false
}

func (f *Frontend) WithId(input string, funcWithId func(id int)) {
	if id := f.SplitCommand(input); id != -1 {
		funcWithId(id)
	}
}

func (f *Frontend) FormatIDs() string {
	jobs := make([]string, 0)
	f.supervisor.ForAllJobs(func(job *JOB.Job) {
		jobName := job.Instances[0].Args[0]
		jobs = append(jobs, fmt.Sprintf("%v: %s\n", job, jobName))
	})
	return strings.Join(jobs, "")
}

func (f *Frontend) FormatJobs() string {
	jobs := make([]string, 0)
	f.supervisor.ForAllJobs(func(job *JOB.Job) {
		for _, instance := range job.Instances {
			if instance.Status != INST.PROCRUNNING || instance.Process == nil {
				continue
			}
			status := instance.GetStatus()
			pid := instance.Process.Pid
			format := "%-12d%-12v%-12d%-12s\n"
			instanceId := instance.InstanceID
			jobString := fmt.Sprintf(format, job.ID, instanceId, pid, status)
			jobs = append(jobs, jobString)
		}
	})
	return strings.Join(jobs, "")
}

func (f *Frontend) SplitCommand(input string) (id int) {
	words := strings.Fields(input)
	switch len(words) {
	case 1:
		id = f.GetId()
	case 2:
		id = f.GetIdFromDefault(words[1])
	default:
		id = -1
	}
	return
}

func (f *Frontend) PrintLogs() {
	data, err := ioutil.ReadFile(f.supervisor.LogFile)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(string(data))
}

func (f *Frontend) GetIdFromDefault(input string) (id int) {
	var err error
	for {
		id, err = strconv.Atoi(input)
		switch {
		case err != nil:
			fallthrough
		case !f.supervisor.HasJob(id):
			fmt.Println("Error: Please enter a valid ID")
			fmt.Print(f.FormatIDs())
			fmt.Print("> ")
			f.scanner.Scan()
			input = f.scanner.Text()
		default:
			return
		}
	}
}

func (f *Frontend) GetId() (id int) {
	var err error
	for {
		fmt.Println("Please select an ID")
		fmt.Print(f.FormatIDs())
		fmt.Print("> ")
		f.scanner.Scan()
		input := f.scanner.Text()
		id, err = strconv.Atoi(input)
		switch {
		case err != nil:
			fallthrough
		case !f.supervisor.HasJob(id):
			fmt.Println("Error: Please enter a valid ID")
		default:
			return
		}
	}
}

func (f *Frontend) PrintHelp() {
	fmt.Println("Commands:")
	fmt.Println("ps:         List current jobs being managed")
	fmt.Println("logs:       display jobs logs")
	fmt.Println("clear:      clear the screen")
	fmt.Println("start [id]: start given job")
	fmt.Println("stop [id]:  stop given job")
	fmt.Println("startAll:   start all jobs")
	fmt.Println("stopAll:    stop all jobs")
	fmt.Println("reload:     reload the configur file")
	fmt.Println("exit:       stop all jobs and exit taskmaster")
}
