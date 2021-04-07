package cmd

import (
	"os"

	"github-release-admin/cmd/create"
	"github-release-admin/cmd/delete"
	"github-release-admin/cmd/list"
	"github-release-admin/log"
)

var (
	List   = list.Run
	Create = create.Run
	Delete = delete.Run
)

func Usage(exitCode int) {
	println(`
a tool for creating, deleting and downloading github release assets.

Usage:
    github-release-admin help <command>
    github-release-admin <repo> [--verbose] <command>

Arguments:
    help                display help message.
    repo                must be specified in the format "owner/repo".

Options:
    --verbose           display verbose output of the execution.

Commands:
    list                list releases.
    create              create release and upload asset files.
	delete              delete releases.

Environment Variables:
    GITHUB_TOKEN        require for private repository
`)
	os.Exit(exitCode)
}

func Help(args []string) {
	if len(args) == 0 {
		Usage(0)
	}

	switch args[0] {
	case "list":
		list.Usage(0)

	case "create":
		create.Usage(0)

	case "delete":
		delete.Usage(0)

	default:
		log.Errorf("invalid command: %q", args[0])
		Usage(1)
	}
}
