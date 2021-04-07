package list

import (
	"encoding/json"
	"os"

	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
)

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

func Usage(exitcode int) {
	println(`
List releases

Usage:
    list [--branch-exist] [--branch=<branch>]
    list draft [--branch-exist] [--branch=<branch>]
    list prerelease [--branch-exist] [--branch=<branch>]

Arguments:
    draft                    lists only the draft releases.
    prelease                 lists only the pre-releases.

Options:
    --branch-exists          lists only the releases associated with the branch
                             that exist.
    --branch=<branch>        lists only the releases associated with the
                             specified branch.
`)
	os.Exit(exitcode)
}

type Option struct {
	BranchExists bool
	Branch       string
}

func (o *Option) SetArg(arg string) bool {
	log.Error("invalid arguments")
	Usage(1)
	return true
}

func (o *Option) SetFlag(arg string) bool {
	switch arg {
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

func display(c *github.Client, o *Option, v *github.Release) {
	if o.Branch != "" && o.Branch != v.TargetCommitish {
		return
	} else if o.BranchExists {
		if b, err := c.GetBranch(v.TargetCommitish); err != nil {
			log.Fatalf("failed to get branch %q: %v", v.TargetCommitish, err)
		} else if b == nil {
			// branch does not exists
			return
		}
	}
	// dump release
	b, _ := json.MarshalIndent(v, "", "  ")
	log.Printf("%s", b)
}

func handleRelease(c *github.Client, o *Option) {
	fetch(c, func(v *github.Release) {
		if v.Draft || v.PreRelease {
			return
		}
		display(c, o, v)
	})
}

func handlePreRelease(c *github.Client, o *Option) {
	fetch(c, func(v *github.Release) {
		if !v.PreRelease {
			return
		}
		display(c, o, v)
	})
}

func handleDraft(c *github.Client, o *Option) {
	fetch(c, func(v *github.Release) {
		if !v.Draft {
			return
		}
		display(c, o, v)
	})
}

func Run(c *github.Client, args []string) {
	o := &Option{}
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	switch arg {
	case "draft":
		getopt.Parse(o, args[1:])
		handleDraft(c, o)

	case "prerelease":
		getopt.Parse(o, args[1:])
		handlePreRelease(c, o)

	default:
		getopt.Parse(o, args)
		handleRelease(c, o)
	}
}
