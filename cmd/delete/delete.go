package delete

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"

	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
)

var osExit = os.Exit

func Usage(code int) {
	println(`
Delete release.

Usage:
    delete <release-id> [--no-dry-run]
    delete by-tag-name <tag> [--regex] [--posix] [--target=<target>]
           [--draft] [--prerelease] [--no-dry-run]

Arguments:
    <release-id>        delete a release with the specified id. (greater than 0)
    by-tag <tag>        delete a release with the specified tag. (e.g. v1.0.0)

Options:
    --regex             compile a tag as regular expressions.
    --posix             compile a tag as POSIX ERE (egrep).
    --target=<target>   specify a branch, or commish. (e.g. master)
    --draft             delete only draft releases.
    --prerelease        delete only prereleases.
    --no-dry-run        actually execute the request.
`)
	osExit(1)
}

type OptionByTag struct {
	ByTag           bool
	TagName         string
	TargetCommitish string
	AsRegex         bool
	AsPosix         bool
	Draft           bool
	PreRelease      bool
	NoDryRun        bool
}

func (o *OptionByTag) SetArg(arg string) bool {
	if o.TagName != "" {
		// <tag> has already passed
		log.Error("invalid arguments")
		Usage(1)

	} else if arg == "" {
		log.Error("invalid <tag> argument")
		Usage(1)
	}
	o.TagName = arg
	return true
}

func (o *OptionByTag) SetFlag(arg string) bool {
	switch arg {
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
		o.NoDryRun = true

	default:
		log.Errorf("unknown option %q", arg)
		Usage(1)
	}

	return true
}

func (o *OptionByTag) SetKeyValue(k, v, arg string) bool {
	switch k {
	case "--target":
		o.TargetCommitish = v

	default:
		log.Errorf("unknown option %q", arg)
		Usage(1)
	}

	return true
}

func fetch(c *github.Client, fn func(*github.Release)) {
	page := 1
	for page > 0 {
		list, err := c.ListReleases(50, page)
		if err != nil {
			log.Fatalf("failed to list releases: %v", err)
		}
		for _, release := range list.Releases {
			fn(release)
		}
		page = list.NextPage
	}
}

func deleteRelease(c *github.Client, v *github.Release, noDryRun bool) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("delete release %d: %s", v.ID, b)
	if noDryRun {
		if err = c.DeleteRelease(v.ID); err != nil {
			log.Fatal(err)
		}
	}
}

func handleDeleteByTagName(c *github.Client, o *OptionByTag) {
	if !o.AsRegex {
		v, err := c.GetReleaseByTagName(o.TagName)
		if err != nil {
			log.Fatalf("failed to get release: %v", err)
		} else if v == nil ||
			(o.Draft && !v.Draft) ||
			(o.PreRelease && !v.PreRelease) ||
			(o.TargetCommitish != "" && v.TargetCommitish != o.TargetCommitish) {
			log.Fatal("release not found")
		}

		deleteRelease(c, v, o.NoDryRun)
		log.Print("OK")
		return
	}

	var re *regexp.Regexp
	var err error
	if o.AsPosix {
		re, err = regexp.CompilePOSIX(o.TagName)
	} else {
		re, err = regexp.Compile(o.TagName)
	}
	if err != nil {
		log.Fatalf("<tag> cannot be compiled as regular expression: %v", err)
	}

	ndelete := 0
	fetch(c, func(v *github.Release) {
		if (o.Draft && !v.Draft) ||
			(o.PreRelease && !v.PreRelease) ||
			(o.TargetCommitish != "" && v.TargetCommitish != o.TargetCommitish) ||
			!re.MatchString(v.TagName) {
			// ignore
			return
		}
		deleteRelease(c, v, o.NoDryRun)
		ndelete++
	})

	if ndelete == 0 {
		log.Fatal("release not found")
	}
	log.Print("OK")
}

type Option struct {
	ReleaseID int64
	NoDryRun  bool
}

func (o *Option) SetArg(arg string) bool {
	if o.ReleaseID != 0 {
		// <release-id> has already passed
		log.Error("invalid arguments")
		Usage(1)
	} else if arg == "" {
		log.Error("invalid <release-id> argument")
		Usage(1)
	}

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

func (o *Option) SetFlag(arg string) bool {
	switch arg {
	case "--no-dry-run":
		o.NoDryRun = true

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

func handleDelete(c *github.Client, o *Option) {
	v, err := c.GetRelease(int(o.ReleaseID))
	if err != nil {
		log.Fatalf("failed to get release: %v", err)
	} else if v == nil {
		log.Fatal("release not found")
	}

	deleteRelease(c, v, o.NoDryRun)
	log.Print("OK")
}

func Run(c *github.Client, args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	switch arg {
	case "by-tag-name":
		o := &OptionByTag{}
		getopt.Parse(o, args[1:])
		if o.TagName == "" {
			log.Error("invalid arguments")
			Usage(1)
		}
		handleDeleteByTagName(c, o)

	default:
		o := &Option{}
		getopt.Parse(o, args)
		if o.ReleaseID == 0 {
			log.Error("invalid arguments")
			Usage(1)
		}
		handleDelete(c, o)
	}
}
