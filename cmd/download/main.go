package main

import (
	"os"
	"strconv"
	"strings"

	"github-release-admin/download"
	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
)

var osExit = os.Exit

func Usage(code int) {
	log.Print(`
Download a release asset.

Usage:
    github-release-download help
    github-release-download <repo> <release-id> <filename> [--verbose]
                            [--no-dry-run]
    github-release-download <repo> latest <filename> [--verbose] [--no-dry-run]
    github-release-download <repo> by-tag <tag>[@<target>] <filename> [--verbose]
                            [--no-dry-run]

Arguments:
    help                display help message.
    <repo>              must be specified in the format "owner/repo".
    <release-id>        dowload from the specified release. (greater than 0)
    <filename>          name of the asset to download.
    latest              download from the lastest release.
    by-tag              download from the release associated with the specified
                        tag (and target).
    <tag>               specify an existing tag. (e.g. v1.0.0)
    <target>            specify a branch, or commish. (e.g. master)

Options:
    --verbose           display verbose output of the execution.
    --no-dry-run        actually execute the request.

Environment Variables:
    GITHUB_TOKEN        required to access the private repository.
`)
	osExit(code)
}

func isNotEmptyString(s string) bool {
	return strings.TrimSpace(s) != ""
}

type Option struct {
	download.Option
	Filename string
}

func (o *Option) SetArg(arg string) bool {
	if o.Filename == "" {
		if isNotEmptyString(arg) {
			o.Filename = arg
			return true
		}
	}
	log.Error("invalid argument")
	Usage(1)
	return false
}

func (o *Option) SetFlag(arg string) bool {
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

func (o *Option) SetKeyValue(k, v, arg string) bool {
	log.Errorf("unknown option %q", arg)
	Usage(1)
	return true
}

type LatestOption struct {
	Option
}

type TagOption struct {
	Option
	TagName         string
	TargetCommitish string
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
		log.Error("invalid arguments")
		Usage(1)
	}

	return o.Option.SetArg(arg)
}

type ReleaseOption struct {
	Option
	ReleaseID int64
}

func (o *ReleaseOption) SetArg(arg string) bool {
	if o.ReleaseID == 0 {
		// verify release-id
		v, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			log.Error("invalid <release-id> argument %w", err)
			Usage(1)
		} else if v <= 0 {
			log.Error("<release-id> must be greater than 0")
			Usage(1)
		}
		o.ReleaseID = v
		return true
	}

	return o.Option.SetArg(arg)
}

func Run(ghc *github.Client, args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	switch arg {
	case "latest":
		o := &LatestOption{}
		o.DryRun = true
		getopt.Parse(o, args[1:])
		if o.Filename == "" {
			log.Error("invalid arguments")
			Usage(1)
		} else if err := download.Latest(
			ghc, o.Filename, &o.Option.Option,
		); err != nil {
			log.Fatalf("failed to download: %v", err)
		}

	case "by-tag":
		o := &TagOption{}
		o.Option.DryRun = true
		getopt.Parse(o, args[1:])
		if o.Filename == "" || o.TagName == "" {
			log.Error("invalid arguments")
			Usage(1)
		} else if err := download.ByTagName(
			ghc, o.TagName, o.TargetCommitish, o.Filename, &o.Option.Option,
		); err != nil {
			log.Fatalf("failed to download: %v", err)
		}

	default:
		o := &ReleaseOption{}
		o.DryRun = true
		getopt.Parse(o, args[1:])
		if o.Filename == "" || o.ReleaseID == 0 {
			log.Error("invalid arguments")
			Usage(1)
		} else if err := download.Release(
			ghc, int(o.ReleaseID), o.Filename, &o.Option.Option,
		); err != nil {
			log.Fatalf("failed to download: %v", err)
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