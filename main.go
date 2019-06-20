package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
)

/*
• The command to use to launch the program
• The number of processes to start and keep running
• Whether to start this program at launch or not
• Whether the program should be restarted always, never, or on unexpected exits
only
• Which return codes represent an "expected" exit status
• How long the program should be running after it’s started for it to be considered
"successfully started"
• How many times a restart should be attempted before aborting
• Which signal should be used to stop (i.e. exit gracefully) the program
• How long to wait after a graceful stop before killing the program
• Options to discard the program’s stdout/stderr or to redirect them to files
• Environment variables to set before launching the program
• A working directory to set before launching the program
• An umask to set before launching the program
*/

var data = `
a: Easy!
b:
  c: 2
  d: [3, 4]
`

// Note: struct fields must be public in order for unmarshal to
// correctly populate the data.
type T struct {
	A string
	B struct {
		RenamedC int   `yaml:"c"`
		D        []int `yaml:",flow"`
	}
}

func main() {
	t := T{}

	err := yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", t)

	d, err := yaml.Marshal(&t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", m)

	d, err = yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))
}
