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

// maybe set conn and err up here

type job struct {
	ID int
	Command string
	Status string
}

var commands map[int]func(pid int64) //create a map for storing commands

const sockAddr = "/tmp/task_master.sock" //Socket address for taskmaster 
// type commandTable func(pid int64)

func init() {
	commands = make(map[int]func(pid int64))
	commands[1] = func (int64){					//stop
		fmt.Println("stop called")
		time.Sleep(1000 * time.Millisecond)
	}
	commands[2] = func (int64){					//start
		fmt.Println("start called")
		time.Sleep(1000 * time.Millisecond)
	}
	commands[3] = func (int64){					//kill
		fmt.Println("kill called")
		time.Sleep(1000 * time.Millisecond)
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
	var jobs []*job // thinking about how to make an array/slice of jobs
	// jobs = append(jobs, newjob)
	// jobs[2]
	conn, err := net.Dial("unix", sockAddr)
	enc := gob.NewEncoder(conn)
	enc.Encode(jobs[2])    // these past three lines were mainly just written to get an idea
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
			fmt.Printf("%-12v%-12v%-12v\n", "ID", "Command", "Status")
			fmt.Print("Please enter a pid: ")
			// line, err := read.ReadString('\n')
			// check pid against vallid pids
			command := commands
			command[cmdNum](1/* actual pid goes here*/)
		}
	}
}
