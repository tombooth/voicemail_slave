package main

import (
	"flag"
	"github.com/tombooth/masterslave"
	"log"
	"os"
	"os/exec"
)

var (
	amqpURI = flag.String("amqpURI", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	execCmd    = flag.String("exec", "cat", "Shell code to be executes")
	wd      = flag.String("wd", "", "Working directory for script")
)

func main() {

	flag.Parse()

	slave, err := masterslave.NewSlave(*amqpURI, "voicemail")

	if err != nil {
		log.Fatalf("Failed to connect masterslave: %s", err)
		os.Exit(1)
	}

	for {
		job := <-slave.Jobs

		log.Println("Got a job to do!")

		cmd := exec.Command(*execCmd)

		if *wd != "" {
			cmd.Dir = *wd
		}

		stdin, err := cmd.StdinPipe()

		if err != nil {
			log.Printf("Failed to get stdin: %s", err)
			job.Failed()
			continue
		}

		numWritten, err := stdin.Write(job.Payload)

		if err != nil {
			log.Printf("Failed to write to stdin: %s", err)
			job.Failed()
			continue
		} else if numWritten < len(job.Payload) {
			log.Printf("Failed to write complete payload")
			job.Failed()
			continue
		}

		stdin.Close()

		outBytes, err := cmd.Output()

		if err != nil {
			log.Println("Failed to run command: %s", err)
			job.Failed()
			continue
		}

		log.Printf("Job succeeded: %s", outBytes)

		job.Done()
	}

}
