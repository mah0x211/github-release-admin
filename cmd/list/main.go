package main

import (
	"encoding/json"
	"os"

	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/list"
	"github-release-admin/log"
)

var osExit = os.Exit

func Usage(code int) {
	log.Print(`
List releases.

Usage:
    github-release-list help
    github-release-list <repo> [--verbose] [--branch-exists] [--branch=<branch>]
    github-release-list <repo> draft [--verbose] [--branch-exists]
                        [--branch=<branch>]
    github-release-list <repo> prerelease [--verbose] [--branch-exists]
                        [--branch=<branch>]

Arguments:
    help                display help message.
    <repo>              must be specified in the format "owner/repo".
    draft               lists only the draft releases.
    prelease            lists only the pre-releases.

Options:
    --verbose           display verbose output of the execution.
    --branch-exists     lists only the releases associated with the branch
                        that exist.
    --branch=<branch>   lists only the releases associated with the
                        specified branch.

Environment Variables:
    GITHUB_TOKEN        required to access the private repository.
`)
	osExit(code)
}

type Option struct {
	list.Option
}

func (o *Option) SetArg(arg string) bool {
	log.Error("invalid arguments")
	Usage(1)
	return true
}

func (o *Option) SetFlag(arg string) bool {
	switch arg {
	case "--verbose":
		log.Verbose = true

	case "--branch-exists":
		o.BranchExists = true

	default:
		log.Errorf("unknown option %q", arg)
		Usage(1)
	}
	return true
}

func (o *Option) SetKeyValue(k, v, arg string) bool {
	switch k {
	case "--branch":
		o.Branch = v

	default:
		log.Errorf("unknown option %q", arg)
		Usage(1)
	}
	return true
}

func Run(ghc *github.Client, args []string) {
	o := &Option{}
	listfn := list.Releases
	if len(args) > 0 {
		switch args[0] {
		case "draft":
			args = args[1:]
			listfn = list.DraftReleases

		case "prerelease":
			args = args[1:]
			listfn = list.PreReleases
		}
	}

	getopt.Parse(o, args)
	v, err := listfn(ghc, &o.Option)
	if err != nil {
		log.Fatalf("failed to list releases: %v", err)
	}

	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("failed to stringify the release list: %v", err)
	}

	log.Print(string(b))
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "help" {
		Usage(0)
	}
	ghc, err := github.New(args[0])
	if err != nil {
		log.Error(err)
		Usage(1)
	}

	Run(ghc, args[1:])
}
