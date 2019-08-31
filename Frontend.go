package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func FormatIDs(s *Supervisor) string {
	jobs := make([]string, 0)
	s.ForAllJobs(func(job *Job) {
		jobs = append(jobs, fmt.Sprintf("%v: %s\n", job, job.Instances[0].args[0]))
	})
	return strings.Join(jobs, "\n") + "\n"
}

func FormatJobs(s *Supervisor) string {
	jobs := make([]string, 0)
	s.ForAllJobs(func(job *Job) {
		for _, instance := range job.Instances {
			if instance.Status == PROCRUNNING {
				jobs = append(jobs, fmt.Sprintf("%d %v %d %d\n", job.ID, instance.InstanceID, instance.state.Pid(), instance.Status))
			} else {
				jobs = append(jobs, fmt.Sprintf("%d %v %d\n", job.ID, instance.InstanceID, instance.Status))
			}
		}
	})
	return strings.Join(jobs, "\n")
}

func StartUI(s *Supervisor) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
Loop:
	for scanner.Scan() {
		cmd := strings.ToLower(scanner.Text())
		switch {
		case cmd == "exit":
			fmt.Println("Exiting TaskMaster")
			s.SigCh <- Signals["SIGTERM"]
			break Loop
		case cmd == "reload":
			fmt.Println("Reloading TaskMaster config")
			s.SigCh <- Signals["SIGHUP"]
		case cmd == "logs":
			data, err := ioutil.ReadFile(s.LogFile)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(data))
		case cmd == "clear":
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
		case cmd == "startAll":
			fmt.Println("Starting all jobs")
			// s.StartAllJobs(false)
		case cmd == "stopAll":
			fmt.Println("Stopping all jobs")
			// s.StopAllJobs(false)
		case strings.HasPrefix(cmd, "start"):
			words := strings.Fields(cmd)
			switch len(words) {
			case 1:
				startLoop := true
				for startLoop {
					fmt.Println("Please select an ID")
					fmt.Println(FormatIDs(s))
					in := scanner.Text()
					id, idErr := strconv.Atoi(in)
					switch {
					case idErr != nil:
						fallthrough
					case !s.HasJob(id):
						fmt.Println("Error: Please enter a valid ID")
					default:
						fmt.Println("Starting", id)
						// s.StartJob(id, false)
						startLoop = false
					}
				}
			case 2:
				startLoop := true
				in := words[1]
				for startLoop {
					id, idErr := strconv.Atoi(in)
					switch {
					case idErr != nil:
						fallthrough
					case !s.HasJob(id):
						fmt.Println("Error: Please enter a valid ID")
						fmt.Println(FormatIDs(s))
						scanner.Scan()
						in = scanner.Text()
					default:
						fmt.Println("Starting", id)
						// s.StartJob(id, false)
						startLoop = false
					}
				}
			}
		case strings.HasPrefix(cmd, "stop"):
			words := strings.Fields(cmd)
			switch len(words) {
			case 1:
				stopLoop := true
				for stopLoop {
					fmt.Println("Please select an ID")
					fmt.Println(FormatIDs(s))
					in := scanner.Text()
					id, idErr := strconv.Atoi(in)
					switch {
					case idErr != nil:
						fallthrough
					case !s.HasJob(id):
						fmt.Println("Error: Please enter a valid ID")
					default:
						fmt.Println("Stopping", id)
						// s.StopJob(id, false)
						stopLoop = false
					}
				}
			case 2:
				stopLoop := true
				in := words[1]
				for stopLoop {
					id, idErr := strconv.Atoi(in)
					switch {
					case idErr != nil:
						fallthrough
					case !s.HasJob(id):
						fmt.Println("Error: Please enter a valid ID")
						fmt.Println(FormatIDs(s))
						scanner.Scan()
						in = scanner.Text()
					default:
						fmt.Println("Stopping", id)
						// s.StopJob(id, false)
						stopLoop = false
					}
				}
			}
		case strings.HasPrefix(cmd, "ps"):
			fmt.Println(FormatJobs(s))
		case strings.HasPrefix(cmd, "help"):
			fmt.Println("Commands:")
			fmt.Println("ps: List current jobs being managed")
			fmt.Println("logs: display jobs logs")
			fmt.Println("clear: clear the screen")
			fmt.Println("start [id]: start given job")
			fmt.Println("stop [id]: stop given job")
			fmt.Println("startAll: start all jobs")
			fmt.Println("stopAll: stop all jobs")
			fmt.Println("reload: reload the configur file")
			fmt.Println("exit: stop all jobs and exit taskmaster")
		}
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			fmt.Fprintln(os.Stderr, err)
			s.SigCh <- Signals["SIGTERM"]
		}
	}
}
