package main

import (
	"context"
	"encoding/json"
	"os"

	"github-release-admin/cmd"
	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/list"
	"github-release-admin/log"
	"github-release-admin/util"
)

var exit = util.Exit

func usage(code int) {
	log.Print(`
List releases.

Usage:
    github-release-list help
    github-release-list [<repo>] [--verbose] [--branch-exists]
                        [--branch=<branch>]
    github-release-list [<repo>] draft [--verbose] [--branch-exists]
                        [--branch=<branch>]
    github-release-list [<repo>] prerelease [--verbose] [--branch-exists]
                        [--branch=<branch>]

Arguments:
    help                display help message.
    <repo>              if the GITHUB_REPOSITORY environment variable is not
                        defined, you must specify the target repository.
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
    GITHUB_REPOSITORY   must be specified in the format "owner/repo".
`)
	exit(code)
}

type Option struct {
	list.Option
}

func (o *Option) SetArg(arg string) bool {
	log.Error("invalid arguments")
	usage(1)
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
		usage(1)
	}
	return true
}

func (o *Option) SetKeyValue(k, v, arg string) bool {
	switch k {
	case "--branch":
		o.Branch = v

	default:
		log.Errorf("unknown option %q", arg)
		usage(1)
	}
	return true
}

func start(ctx context.Context, ghc *github.Client, args []string) {
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
	os.Exit(cmd.Start(start, usage))
}
