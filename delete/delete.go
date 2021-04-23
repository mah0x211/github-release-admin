package delete

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github-release-admin/github"
	"github-release-admin/log"
)

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

var ErrNotFound = fmt.Errorf("not found")

type UnbranchedReleasesOption struct {
	ItemsPerPage int
	DryRun       bool
}

func UnbranchedReleases(ghc *github.Client, o *UnbranchedReleasesOption) error {
	ndelete := 0
	if err := ghc.FetchRelease(1, o.ItemsPerPage, func(v *github.Release, _ int) error {
		if b, err := ghc.GetBranch(v.TargetCommitish); err != nil {
			return err
		} else if b != nil {
			// branch exists
			log.Debug("ignore the release associated with the branch: %d", v.ID)
			return nil
		} else if err = deleteRelease(ghc, v, o.DryRun); err != nil {
			return err
		}
		ndelete++
		return nil
	}); err != nil {
		return err
	}

	if ndelete == 0 {
		return ErrNotFound
	}
	return nil
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

func ReleasesByTagName(ghc *github.Client, o *ReleasesByTagNameOption) error {
	if !o.AsRegex {
		v, err := ghc.GetReleaseByTagName(o.TagName)
		if err != nil {
			return err
		} else if v == nil || !isDeletionTarget(v, o, nil) {
			return ErrNotFound
		}
		return deleteRelease(ghc, v, o.DryRun)
	}

	var re *regexp.Regexp
	var err error
	if o.AsPosix {
		re, err = regexp.CompilePOSIX(o.TagName)
	} else {
		re, err = regexp.Compile(o.TagName)
	}
	if err != nil {
		return fmt.Errorf(
			"%q cannot be compiled as regular expression: %w", o.TagName, err,
		)
	}

	ndelete := 0
	if err = ghc.FetchRelease(1, o.ItemsPerPage, func(v *github.Release, _ int) error {
		if !isDeletionTarget(v, o, re) {
			return nil
		} else if err = deleteRelease(ghc, v, o.DryRun); err != nil {
			return err
		}
		ndelete++
		return nil
	}); err != nil {
		return err
	}

	if ndelete == 0 {
		return ErrNotFound
	}
	return nil
}

type ReleaseOption struct {
	ReleaseID int64
	DryRun    bool
}

func Release(ghc *github.Client, o *ReleaseOption) error {
	v, err := ghc.GetRelease(int(o.ReleaseID))
	if err != nil {
		return err
	} else if v == nil {
		return ErrNotFound
	}

	return deleteRelease(ghc, v, o.DryRun)
}
