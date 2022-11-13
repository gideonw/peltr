package main

import (
	"os"

	"github.com/gideonw/peltr/cmd"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	// Example of verbosity with level
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
}

var options Options

var parser = flags.NewParser(&options, flags.Default)

func init() {
	parser.AddCommand("server",
		"Coordination and data collection server",
		"Server to coordinate jobs and collect data on the jobs that are in progress.",
		&cmd.ServerCmd)
	parser.AddCommand("worker",
		"Job running worker",
		"Worker to run and process jobs and send metrics collected to the server",
		&cmd.WorkerCmd)
}

func main() {
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}
