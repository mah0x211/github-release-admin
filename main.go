package main

import (
	"os"

	"github-release-admin/cmd"
	"github-release-admin/github"
	"github-release-admin/log"
)

func parseCommand(c *github.Client, args []string) {
	if len(args) == 0 {
		log.Error("no command argument")
		cmd.Usage(1)
	}

	switch args[0] {
	case "list":
		cmd.List(c, args[1:])

	case "create":
		cmd.Create(c, args[1:])

	default:
		log.Errorf("invalid command %q", args[0])
		cmd.Usage(1)
	}
}

func parseFlags(c *github.Client, args []string) {
	if len(args) > 0 && args[0] == "--verbose" {
		log.Verbose = true
		args = args[1:]
	}
	parseCommand(c, args)
}

func parseRepo(args []string) {
	if len(args) == 0 {
		cmd.Usage(0)
	} else if args[0] == "help" {
		cmd.Help(args[1:])
	}

	c, err := github.New(args[0])
	if err != nil {
		log.Error(err)
		cmd.Usage(1)
	}

	parseFlags(c, args[1:])
}

func main() {
	parseRepo(os.Args[1:])
}
