package main

import (
	"os"
	"strconv"
	"strings"

	"github-release-admin/delete"
	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
)

var osExit = os.Exit

func Usage(code int) {
	log.Print(`
Delete release.

Usage:
    github-release-delete help
    github-release-delete <repo> <release-id> [--verbose] [--no-dry-run]
    github-release-delete <repo> no-branch [--verbose] [--no-dry-run]
    github-release-delete <repo> by-tag <tag>[@<target>] [--verbose] [--no-dry-run]
                          [--regex] [--posix] [--draft] [--prerelease]

Arguments:
    help                display help message.
    <repo>              must be specified in the format "owner/repo".
    <release-id>        delete a release with the specified id. (greater than 0)
    by-tag              delete a release with the specified tag.
    <tag>               specify an existing tag. (e.g. v1.0.0)
    <target>            specify a branch, or commish. (e.g. master)

Options:
    --verbose           display verbose output of the execution.
    --regex             compile a <tag> as regular expressions.
    --posix             compile a <tag> as POSIX ERE (egrep).
    --draft             delete only draft releases.
    --prerelease        delete only prereleases.
    --no-dry-run        actually execute the request.

Environment Variables:
    GITHUB_TOKEN        required to access the private repository.
`)
	osExit(code)
}

func isNotEmptyString(s string) bool {
	return strings.TrimSpace(s) != ""
}

type NoBranchOption struct {
	delete.NoBranchOption
}

func (o *NoBranchOption) SetArg(arg string) bool {
	log.Error("invalid arguments")
	Usage(1)
	return true
}

func (o *NoBranchOption) SetFlag(arg string) bool {
	switch arg {
	case "--verbose":
		log.Verbose = true

	case "--no-dry-run":
		o.DryRun = false

	default:
		log.Errorf("unknown option %q", arg)
		Usage(1)
	}

	return true
}

func (o *NoBranchOption) SetKeyValue(k, v, arg string) bool {
	log.Errorf("unknown option %q", arg)
	Usage(1)
	return true
}

type TagOption struct {
	delete.TagOption
}

func (o *TagOption) SetArg(arg string) bool {
	if o.TagName == "" {
		arr := strings.Split(arg, "@")
		switch len(arr) {
		case 1:
			if isNotEmptyString(arr[0]) {
				o.TagName = arr[0]
				return true
			}

		case 2:
			if isNotEmptyString(arr[0]) && isNotEmptyString(arr[1]) {
				o.TagName = arr[0]
				o.TargetCommitish = arr[1]
				return true
			}
		}
		log.Error("invalid <tag>[@<target>] arguments")
		Usage(1)
	}

	log.Error("invalid arguments")
	Usage(1)
	return true
}

func (o *TagOption) SetFlag(arg string) bool {
	switch arg {
	case "--verbose":
		log.Verbose = true

	case "--posix":
		o.AsPosix = true
		fallthrough
	case "--regex":
		o.AsRegex = true

	case "--draft":
		o.Draft = true

	case "--prerelease":
		o.PreRelease = true

	case "--no-dry-run":
		o.DryRun = false

	default:
		log.Errorf("unknown option %q", arg)
		Usage(1)
	}

	return true
}

func (o *TagOption) SetKeyValue(k, v, arg string) bool {
	log.Errorf("unknown option %q", arg)
	Usage(1)
	return true
}

type ReleaseOption struct {
	delete.ReleaseOption
}

func (o *ReleaseOption) SetArg(arg string) bool {
	if o.ReleaseID == 0 {
		if isNotEmptyString(arg) {
			// verify release-id
			if v, err := strconv.ParseInt(arg, 10, 64); err == nil && v > 0 {
				o.ReleaseID = v
				return true
			}
		}

		log.Error("invalid <release-id> argument")
		Usage(1)
	}

	// <release-id> has already passed
	log.Error("invalid arguments")
	Usage(1)
	return true
}

func (o *ReleaseOption) SetFlag(arg string) bool {
	switch arg {
	case "--verbose":
		log.Verbose = true

	case "--no-dry-run":
		o.DryRun = false

	default:
		log.Errorf("unknown option %q", arg)
		Usage(1)
	}

	return true
}

func (o *ReleaseOption) SetKeyValue(k, v, arg string) bool {
	log.Errorf("unknown option %q", arg)
	Usage(1)
	return true
}

func Run(ghc *github.Client, args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	switch arg {
	case "no-branch":
		o := &NoBranchOption{}
		o.DryRun = true

		getopt.Parse(o, args[1:])
		if err := delete.NoBranch(ghc, &o.NoBranchOption); err != nil {
			log.Fatalf("failed to delete release: %v", err)
		}

	case "by-tag":
		o := &TagOption{}
		o.DryRun = true

		getopt.Parse(o, args[1:])
		if o.TagName == "" {
			log.Error("invalid arguments")
			Usage(1)
		} else if err := delete.ByTagName(ghc, &o.TagOption); err != nil {
			log.Fatalf("failed to delete release: %v", err)
		}

	default:
		o := &ReleaseOption{}
		o.DryRun = true

		getopt.Parse(o, args)
		if o.ReleaseID == 0 {
			log.Error("invalid arguments")
			Usage(1)
		} else if err := delete.Release(ghc, &o.ReleaseOption); err != nil {
			log.Fatalf("failed to delete release: %v", err)
		}
	}
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
