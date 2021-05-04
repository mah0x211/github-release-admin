package delete

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/mah0x211/github-release-admin/github"
	"github.com/mah0x211/github-release-admin/log"
)

func deleteTag(ghc *github.Client, v *github.Release, dryrun bool) error {
	if log.Verbose {
		log.Debug("delete tag %s", v.TagName)
	}

	if dryrun {
		return nil
	}
	return ghc.DeleteTag(v.TagName)
}

func deleteRelease(ghc *github.Client, v *github.Release, dryrun bool) error {
	if log.Verbose {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		log.Debug("delete release %d: %s", v.ID, b)
	}

	if dryrun {
		return nil
	}
	return ghc.DeleteRelease(v.ID)
}

var reHex = regexp.MustCompile("^[0-9a-fA-F]+$")

func IsHexSHA1(s string) bool {
	return len(s) == 40 && reHex.MatchString(s)
}

func getBranches(ghc *github.Client, targetCommitish string) ([]*github.Branch, error) {
	if IsHexSHA1(targetCommitish) {
		return ghc.ListBranchesOfCommit(targetCommitish, 20, 20)
	} else if v, err := ghc.GetBranch(targetCommitish); err != nil {
		return nil, err
	} else if v != nil {
		return []*github.Branch{v}, nil
	}
	return nil, nil
}

type UnbranchedReleasesOption struct {
	ItemsPerPage int
	DryRun       bool
}

func UnbranchedReleases(ghc *github.Client, o *UnbranchedReleasesOption) ([]*github.Release, error) {
	list := []*github.Release{}
	if err := ghc.FetchRelease(1, o.ItemsPerPage, func(v *github.Release, _ int) error {
		if list, err := getBranches(ghc, v.TargetCommitish); err != nil {
			return err
		} else if len(list) > 0 {
			log.Debug("ignore the release associated with the branch: %d", v.ID)
			return nil
		} else if err = deleteRelease(ghc, v, o.DryRun); err != nil {
			return err
		}
		list = append(list, v)
		return deleteTag(ghc, v, o.DryRun)
	}); err != nil {
		return list, err
	}

	return list, nil
}

type DraftReleasesOption struct {
	ItemsPerPage int
	DryRun       bool
	Branch       string
}

func DraftReleases(ghc *github.Client, o *DraftReleasesOption) ([]*github.Release, error) {
	list := []*github.Release{}
	if err := ghc.FetchRelease(1, o.ItemsPerPage, func(v *github.Release, _ int) error {
		if !v.Draft {
			log.Debug("ignore non-draft release: %d", v.ID)
			return nil
		} else if o.Branch != "" && v.TargetCommitish != o.Branch {
			log.Debug("ignore release that branch does not matched to %q: %d", o.Branch, v.ID)
			return nil
		} else if err := deleteRelease(ghc, v, o.DryRun); err != nil {
			return err
		}
		list = append(list, v)
		return deleteTag(ghc, v, o.DryRun)
	}); err != nil {
		return list, err
	}

	return list, nil
}

type PreReleasesOption struct {
	ItemsPerPage int
	DryRun       bool
	Branch       string
}

func PreReleases(ghc *github.Client, o *PreReleasesOption) ([]*github.Release, error) {
	list := []*github.Release{}
	if err := ghc.FetchRelease(1, o.ItemsPerPage, func(v *github.Release, _ int) error {
		if !v.PreRelease {
			log.Debug("ignore non-prerelease: %d", v.ID)
			return nil
		} else if o.Branch != "" && v.TargetCommitish != o.Branch {
			log.Debug("ignore release that branch does not matched to %q: %d", o.Branch, v.ID)
			return nil
		} else if err := deleteRelease(ghc, v, o.DryRun); err != nil {
			return err
		}
		list = append(list, v)
		return deleteTag(ghc, v, o.DryRun)
	}); err != nil {
		return list, err
	}

	return list, nil
}

type ReleasesByTagNameOption struct {
	ItemsPerPage    int
	TagName         string
	TargetCommitish string
	AsRegex         bool
	AsPosix         bool
	Draft           bool
	PreRelease      bool
	DryRun          bool
}

func isDeletionTarget(v *github.Release, o *ReleasesByTagNameOption, re *regexp.Regexp) bool {
	if o.Draft && !v.Draft {
		log.Debug("ignore non-draft release: %d", v.ID)
		return false
	} else if o.PreRelease && !v.PreRelease {
		log.Debug("ignore non-prerelease: %d", v.ID)
		return false
	} else if o.TargetCommitish != "" && v.TargetCommitish != o.TargetCommitish {
		log.Debug("ignore release that commitish does not matched to %q: %d", o.TargetCommitish, v.ID)
		return false
	} else if re != nil && !re.MatchString(v.TagName) {
		log.Debug("ignore release that tag-name does not matched to %q: %d", o.TagName, v.ID)
		return false
	}
	return true
}

func ReleasesByTagName(ghc *github.Client, o *ReleasesByTagNameOption) ([]*github.Release, error) {
	list := []*github.Release{}

	if !o.AsRegex {
		v, err := ghc.GetReleaseByTagName(o.TagName)
		if err != nil {
			return list, err
		} else if v == nil || !isDeletionTarget(v, o, nil) {
			return list, nil
		} else if err = deleteRelease(ghc, v, o.DryRun); err != nil {
			return list, err
		}
		return append(list, v), deleteTag(ghc, v, o.DryRun)
	}

	var re *regexp.Regexp
	var err error
	if o.AsPosix {
		re, err = regexp.CompilePOSIX(o.TagName)
	} else {
		re, err = regexp.Compile(o.TagName)
	}
	if err != nil {
		return list, fmt.Errorf(
			"%q cannot be compiled as regular expression: %w", o.TagName, err,
		)
	}

	if err = ghc.FetchRelease(1, o.ItemsPerPage, func(v *github.Release, _ int) error {
		if !isDeletionTarget(v, o, re) {
			return nil
		} else if err = deleteRelease(ghc, v, o.DryRun); err != nil {
			return err
		}
		list = append(list, v)
		return deleteTag(ghc, v, o.DryRun)
	}); err != nil {
		return list, err
	}

	return list, nil
}

type ReleaseOption struct {
	ReleaseID int64
	DryRun    bool
}

func Release(ghc *github.Client, o *ReleaseOption) (*github.Release, error) {
	v, err := ghc.GetRelease(int(o.ReleaseID))
	if err != nil {
		return nil, err
	} else if v == nil {
		return nil, nil
	} else if err = deleteRelease(ghc, v, o.DryRun); err != nil {
		return nil, err
	}
	return v, deleteTag(ghc, v, o.DryRun)
}
