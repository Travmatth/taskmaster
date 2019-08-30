package main

import (
	"fmt"
	"encoding/gob"
	"net"
	"os"
	"log"
	"time"
	"os/exec"
)

const sockAddr = "/tmp/task_master.sock"

type sJob struct {
	ID int
	Command string
	Status string
}

func clearwindow() {
	cmd := exec.Command("clear") //function to run clear on mac/linux
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	var job sJob
	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatal(err)
	}
	clearwindow()
	l, err := net.Listen("unix", sockAddr)
	if err != nil {
		fmt.Printf("Error: could not listen on %s\n", sockAddr)
		return
	}
	fmt.Printf("Connected to taskmaster shell on %s\n", sockAddr)
	defer l.Close()

	a, _ := l.Accept()
	dec := gob.NewDecoder(a)
	for ;true; {
		err = dec.Decode(&job)
		if err != nil {
			clearwindow()
			fmt.Print("Error: could not decode job\n")
		} else {
			clearwindow()
			fmt.Println(job)
			time.Sleep(1000 * time.Millisecond)
		}
	}
	l.Close()
}
