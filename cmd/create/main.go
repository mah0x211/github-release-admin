package main

import (
	"os"
	"regexp"
	"strings"

	"github-release-admin/create"
	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
	"github-release-admin/readdir"
)

var osExit = os.Exit

func Usage(code int) {
	log.Print(`
Create release and upload asset files.

Usage:
    github-release-create help
    github-release-create <repo> <tag>[@<target>] <filename>
           [--verbose] [--title=<title>] [--body=<body>]
           [--dir=<path/to/dir>] [--regex] [--posix]
           [--no-draft] [--no-prerelease] [--no-dry-run]

Arguments:
    help                display help message.
    <repo>              must be specified in the format "owner/repo".
    <tag>               specify an existing tag, or create a new tag.
                        (e.g. v1.0.0)
    <target>            specify a branch, or commish. (e.g. master)
    <filename>          name of the asset file to upload. (e.g. myasset.tar.gz)

Options:
    --verbose           display verbose output of the execution.
    --title=<title>     release title.
    --body=<body>       describe this release.
    --dir=<path/to/dir> reads the file from this directory.
    --regex             compile <filename> as regular expressions.
    --posix             compile <filename> as POSIX ERE (egrep).
    --no-draft          save as non-draft release.
    --no-prerelease     save as non-prerelease (production ready).
    --no-dry-run        actually execute the request.

Environment Variables:
    GITHUB_TOKEN        required to access the private repository.
`)
	osExit(code)
}

type Option struct {
	create.Option
}

func isNotEmptyString(s string) bool {
	return strings.TrimSpace(s) != ""
}

func (o *Option) SetArg(arg string) bool {
	if o.TagName == "" {
		// parse <tag>[@<target>]
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

	} else if o.Filename == "" {
		// parse <filename>
		if isNotEmptyString(arg) {
			o.Filename = arg
			return true
		}
		log.Error("invalid <filename> arguments")
		Usage(1)
	} else {
		log.Error("invalid arguments")
		Usage(1)
	}

	return true
}

func (o *Option) SetFlag(arg string) bool {
	switch arg {
	case "--verbose":
		log.Verbose = true

	case "--posix":
		o.AsPosix = true
		fallthrough
	case "--regex":
		o.AsRegex = true

	case "--no-draft":
		o.Draft = false

	case "--no-prerelease":
		o.PreRelease = false

	case "--no-dry-run":
		o.DryRun = false

	default:
		log.Errorf("unknown option %q", arg)
		Usage(1)
	}

	return true
}

func (o *Option) SetKeyValue(k, v, arg string) bool {
	switch k {
	case "--title":
		o.Title = v

	case "--body":
		o.Body = v

	case "--dir":
		o.Dirname = v

	default:
		log.Errorf("unknown option %q", arg)
		Usage(1)
	}
	return true
}

func Run(ghc *github.Client, args []string) {
	o := &Option{}
	o.Draft = true
	o.PreRelease = true
	o.DryRun = true
	getopt.Parse(o, args)

	if o.TagName == "" || o.Filename == "" {
		log.Error("invalid arguments")
		Usage(1)
	}

	var re *regexp.Regexp
	var err error
	if o.AsPosix {
		re, err = regexp.CompilePOSIX(o.Filename)
	} else if o.AsRegex {
		re, err = regexp.Compile(o.Filename)
	}
	if err != nil {
		log.Errorf(
			"<filename> cannot be compiled as regular expressions: %v", err,
		)
		Usage(1)
	}

	// read asset files
	r := readdir.New(o.Dirname, o.Filename, re)
	assets, err := r.Read()
	if err != nil {
		log.Fatalf("failed to readdir(): %v", err)
	} else if len(assets) == 0 {
		log.Print("asset files not found")
		return
	}

	if err = create.Release(ghc, &o.Option, r, assets); err != nil {
		log.Fatalf("failed to create release: %v", err)
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
