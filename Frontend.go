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
	return strings.Join(jobs, "") + "\n"
}

func FormatJobs(s *Supervisor) string {
	jobs := make([]string, 0)
	var pid int
	fmt.Printf("%-12v%-12v%-12v%-12v\n", "ID", "Instance", "PID", "Status")
	s.ForAllJobs(func(job *Job) {
		for _, instance := range job.Instances {
			status := instance.GetStatus()
			if instance.process != nil {
				pid = instance.GetPid()
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
						s.StopJob(id, false)
						stopLoop = false
						fmt.Print("> ")
					}
				}
			}
		case strings.HasPrefix(cmd, "ps"):
			fmt.Println(FormatJobs(s))
			fmt.Print("> ")
		case strings.HasPrefix(cmd, "help"):
			fmt.Println("Commands:")
			fmt.Printf("%-14s %s\n", "ps:", " List current jobs being managed")
			fmt.Printf("%-14s %s\n", "logs:", " display jobs logs")
			fmt.Printf("%-14s %s\n", "clear:", " clear the screen")
			fmt.Printf("%-14s %s\n", "start[id]: ", " start given job")
			fmt.Printf("%-14s %s\n", "stop[id]: ", " stop given job")
			fmt.Printf("%-14s %s\n", "startAll:", " start all jobs")
			fmt.Printf("%-14s %s\n", "stopAll:", " stop all jobs")
			fmt.Printf("%-14s %s\n", "reload:", " reload the configur file")
			fmt.Printf("%-14s %s\n> ", "exit:", " stop all jobs and exit taskmaster")
		default:
			fmt.Print("Error: unrecognized command\n> ")
		}
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			fmt.Fprintln(os.Stderr, err)
			s.SigCh <- Signals["SIGTERM"]
		}
	}
}
