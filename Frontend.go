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
	return strings.Join(jobs, "")
}

func FormatJobs(s *Supervisor) string {
	jobs := make([]string, 0)
	s.ForAllJobs(func(job *Job) {
		for _, instance := range job.Instances {
			var pid int
			status := instance.GetStatus()
			if instance.Status == PROCRUNNING && instance.process != nil {
				pid = instance.process.Pid
			} else {
				pid = -1
			}
			jobs = append(jobs, fmt.Sprintf("%-12d%-12v%-12d%-12s\n", job.ID, instance.InstanceID, pid, status))
		}
	})
	return strings.Join(jobs, "")
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
				fmt.Print(err)
			}
			fmt.Print(string(data))
		case cmd == "clear":
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
		case cmd == "startall" || cmd == "start all":
			fmt.Println("Starting all jobs")
			s.StartAllJobs(false)
		case cmd == "stopall" || cmd == "stop all":
			fmt.Println("Stopping all jobs")
			s.StopAllJobs(false)
		case strings.HasPrefix(cmd, "start"):
			words := strings.Fields(cmd)
			switch len(words) {
			case 1:
				startLoop := true
				for startLoop {
					fmt.Println("Please select an ID")
					fmt.Print(FormatIDs(s))
					fmt.Print("> ")
					scanner.Scan()
					in := scanner.Text()
					id, idErr := strconv.Atoi(in)
					switch {
					case idErr != nil:
						fallthrough
					case !s.HasJob(id):
						fmt.Println("Error: Please enter a valid ID")
					default:
						fmt.Println("Starting", id)
						s.StartJob(id, false)
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
						fmt.Print(FormatIDs(s))
						fmt.Print("> ")
						scanner.Scan()
						in = scanner.Text()
					default:
						fmt.Println("Starting", id)
						s.StartJob(id, false)
						startLoop = false
					}
				}
			}
		case strings.HasPrefix(cmd, "restart"):
			words := strings.Fields(cmd)
			switch len(words) {
			case 1:
				startLoop := true
				for startLoop {
					fmt.Println("Please select an ID")
					fmt.Print(FormatIDs(s))
					fmt.Print("> ")
					scanner.Scan()
					in := scanner.Text()
					id, idErr := strconv.Atoi(in)
					switch {
					case idErr != nil:
						fallthrough
					case !s.HasJob(id):
						fmt.Println("Error: Please enter a valid ID")
					default:
						fmt.Println("Restarting", id)
						s.RestartJob(id, false)
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
						fmt.Print(FormatIDs(s))
						fmt.Print("> ")
						scanner.Scan()
						in = scanner.Text()
					default:
						fmt.Println("Restarting", id)
						s.RestartJob(id, false)
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
					fmt.Print(FormatIDs(s))
					fmt.Print("> ")
					scanner.Scan()
					in := scanner.Text()
					id, idErr := strconv.Atoi(in)
					switch {
					case idErr != nil:
						fallthrough
					case !s.HasJob(id):
						fmt.Println("Error: Please enter a valid ID")
					default:
						fmt.Println("Stopping", id)
						s.StopJob(id)
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
						fmt.Print(FormatIDs(s))
						fmt.Print("> ")
						scanner.Scan()
						in = scanner.Text()
					default:
						fmt.Println("Stopping", id)
						s.StopJob(id)
						stopLoop = false
					}
				}
			}
		case strings.HasPrefix(cmd, "ps"):
			fmt.Printf("%-12s%-12s%-12s%-12s\n", "ID", "Instance", "PID", "Status")
			fmt.Print(FormatJobs(s))
		case strings.HasPrefix(cmd, "help"):
			fmt.Println("Commands:")
			fmt.Println("ps:         List current jobs being managed")
			fmt.Println("logs:       display jobs logs")
			fmt.Println("clear:      clear the screen")
			fmt.Println("start [id]: start given job")
			fmt.Println("restart [id]: restart given job")
			fmt.Println("stop [id]:  stop given job")
			fmt.Println("startAll:   start all jobs")
			fmt.Println("stopAll:    stop all jobs")
			fmt.Println("reload:     reload the configur file")
			fmt.Println("exit:       stop all jobs and exit taskmaster")
		}
		fmt.Print("> ")
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			fmt.Fprintln(os.Stderr, err)
			s.SigCh <- Signals["SIGTERM"]
		}
	}
}
