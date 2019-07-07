package main

import (
	"fmt"
	"bufio"
	"strings"
	"strconv"
	"encoding/gob"
	// "errors"
	"os"
	"os/exec"
	"time"
	"net"
)

var conn net.Conn
var err error

type sJob struct {
	ID int
	Command string
	Status string
}

var commands map[int]func(pid int64, job sJob, enc *gob.Encoder) (sJob) //create a map for storing commands

const sockAddr = "/tmp/task_master.sock" //Socket address for taskmaster 
// type commandTable func(pid int64)

func init() {
	commands = make(map[int]func(pid int64, job sJob, enc *gob.Encoder)(sJob))
	commands[1] = func (pid int64, job sJob, enc *gob.Encoder) (sJob) {				//stop
		fmt.Println("stop called in client")
		time.Sleep(1000 * time.Millisecond)
		job.ID = 1
		job.Command = "stop called in server"
		job.Status = "stopped"
		enc.Encode(job)
		return job
	}
	commands[2] = func (pid int64, job sJob, enc *gob.Encoder) (sJob) {				//start
		fmt.Println("start called in client")
		time.Sleep(1000 * time.Millisecond)
		job.ID = 2
		job.Command = "start called in server"
		job.Status = "started"
		enc.Encode(job)
		return job
	}
	commands[3] = func (pid int64, job sJob, enc *gob.Encoder) (sJob) {				//kill
		fmt.Println("kill called in client")
		time.Sleep(1000 * time.Millisecond)
		job.ID = 3
		job.Command = "kill called in server"
		job.Status = "killed"
		enc.Encode(job)
		return job
	}
}

func clearwindow() {
	cmd := exec.Command("clear") //function to run clear on mac/linux
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func validInputCheck(n int) (bool) {
	return (n == 1 || n == 2 || n == 3 || n == 4)
}

func main() {
	var job sJob
	// var jobs []sJob // thinking about how to make an array/slice of jobs
	// jobs = append(jobs, newjob)
	// jobs[2]
	//start taskmaster server/daemon if no started
	conn, err = net.Dial("unix", sockAddr)
	if err != nil {
		cmd := exec.Command("go run server/test_server.go &")
		cmd.Stdout = os.Stdout
		cmd.Run()
		conn, err = net.Dial("unix", sockAddr)
		if err != nil {
			fmt.Printf("Error: unable to connect to taskmaster daemon at %s\n", sockAddr)
			time.Sleep(1000 * time.Millisecond)
			return
		}
		fmt.Printf("Connected to taskmaster daemon at %s\n", sockAddr)
	}
	enc := gob.NewEncoder(conn)
	for ;true;{
		var cmdNum int
		clearwindow()
		fmt.Printf("%-12v%-12v%-12v\n", "ID", "Command", "Status")
		// for i = 0; i < pids; i++{				this is for when we have a pid to give it
			// fmt.Println("%-12v%-12v%-12v", , ,)
			// }
		fmt.Printf("Commands: %v, %v, %v\n\n", "1:start", "2:stop", "3:kill")
		fmt.Print("Please choose a command number: ")
		read := bufio.NewReader(os.Stdin)
		
		line, err := read.ReadString('\n')
		
		cmdNum, _ = strconv.Atoi(strings.TrimSuffix(line, "\n"))
		// fmt.Printf("%d\n", cmdNum)
		// time.Sleep(1000 * time.Millisecond)
		if err != nil || !validInputCheck(cmdNum) {
			fmt.Println("not a valid input")
			time.Sleep(500 * time.Millisecond)
		} else if cmdNum == 4 {
			clearwindow()
			break
		} else {
			clearwindow()
			// jobs[0] = append()
			// fmt.Printf("%-12v%-12v%-12v\n", "ID", "Command", "Status")
			// fmt.Print("Please enter a pid: ")
			// line, err := read.ReadString('\n')
			// check pid against vallid pids
			command := commands
			_ = command[cmdNum](1, job /* actual pid goes here*/, enc)
		}
	}
}
