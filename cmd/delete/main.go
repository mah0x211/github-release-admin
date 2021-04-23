package main

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github-release-admin/cmd"
	"github-release-admin/delete"
	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
	"github-release-admin/util"
)

var exit = util.Exit

func usage(code int) {
	log.Print(`
Delete release.

Usage:
    github-release-delete help
    github-release-delete [<repo>] <release-id> [--verbose] [--no-dry-run]
    github-release-delete [<repo>] unbranched [--verbose] [--no-dry-run]
    github-release-delete [<repo>] by-tag <tag>[@<target>] [--verbose]
                          [--no-dry-run] [--regex] [--posix] [--draft]
                          [--prerelease]

Arguments:
    help                display help message.
    <repo>              if the GITHUB_REPOSITORY environment variable is not
                        defined, you must specify the target repository.
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
    GITHUB_REPOSITORY   must be specified in the format "owner/repo".
`)
	exit(code)
}

func isNotEmptyString(s string) bool {
	return strings.TrimSpace(s) != ""
}

type UnbranchedReleasesOption struct {
	delete.UnbranchedReleasesOption
}

func (o *UnbranchedReleasesOption) SetArg(arg string) bool {
	log.Error("invalid arguments")
	usage(1)
	return true
}

func (o *UnbranchedReleasesOption) SetFlag(arg string) bool {
	switch arg {
	case "--verbose":
		log.Verbose = true

	case "--no-dry-run":
		o.DryRun = false

	default:
		log.Errorf("unknown option %q", arg)
		usage(1)
	}

	return true
}

func (o *UnbranchedReleasesOption) SetKeyValue(k, v, arg string) bool {
	log.Errorf("unknown option %q", arg)
	usage(1)
	return true
}

type ReleasesByTagNameOption struct {
	delete.ReleasesByTagNameOption
}

func (o *ReleasesByTagNameOption) SetArg(arg string) bool {
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
		usage(1)
	}

	log.Error("invalid arguments")
	usage(1)
	return true
}

func (o *ReleasesByTagNameOption) SetFlag(arg string) bool {
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
		usage(1)
	}

	return true
}

func (o *ReleasesByTagNameOption) SetKeyValue(k, v, arg string) bool {
	log.Errorf("unknown option %q", arg)
	usage(1)
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
		usage(1)
	}

	// <release-id> has already passed
	log.Error("invalid arguments")
	usage(1)
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
		usage(1)
	}

	return true
}

func (o *ReleaseOption) SetKeyValue(k, v, arg string) bool {
	log.Errorf("unknown option %q", arg)
	usage(1)
	return true
}

func start(ctx context.Context, ghc *github.Client, args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	var list []*github.Release
	var err error

	switch arg {
	case "unbranched":
		o := &UnbranchedReleasesOption{}
		o.DryRun = true
		getopt.Parse(o, args[1:])
		list, err = delete.UnbranchedReleases(ghc, &o.UnbranchedReleasesOption)

	case "by-tag":
		o := &ReleasesByTagNameOption{}
		o.DryRun = true

		getopt.Parse(o, args[1:])
		if o.TagName == "" {
			log.Error("invalid arguments")
			usage(1)
		} else if err := delete.ReleasesByTagName(
			ghc, &o.ReleasesByTagNameOption,
		); err != nil {
			log.Fatalf("failed to delete release: %v", err)
		}

	default:
		o := &ReleaseOption{}
		o.DryRun = true

		getopt.Parse(o, args)
		if o.ReleaseID == 0 {
			log.Error("invalid arguments")
			usage(1)
		} else if err := delete.Release(ghc, &o.ReleaseOption); err != nil {
			log.Fatalf("failed to delete release: %v", err)
		}
	}

	b, _ := json.MarshalIndent(list, "", "  ")
	log.Print(string(b))
	if err != nil {
		log.Fatalf("failed to delete release: %v", err)
	}
}

func main() {
	os.Exit(cmd.Start(start, usage))
}
