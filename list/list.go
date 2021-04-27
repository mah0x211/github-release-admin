package list

import (
	"errors"

	"github.com/mah0x211/github-release-admin/github"
	"github.com/mah0x211/github-release-admin/log"
)

type Option struct {
	ItemsPerPage int
	MaxItems     uint64
	BranchExists bool
	Branch       string
}

const (
	flgReleaseOnly  = 0x0
	flgDraftRelease = 0x1
	flgPreRelease   = 0x2
	flgAll          = 0x3
)

func isListTarget(v *github.Release, flg int, o *Option) bool {
	if flg == flgReleaseOnly {
		if v.Draft || v.PreRelease {
			log.Debug("ignore draft or prerelease: %d", v.ID)
			return false
		}
	} else if flg != flgAll {
		if flg&flgDraftRelease != 0 && !v.Draft {
			log.Debug("ignore non-draft release: %d", v.ID)
			return false
		}
		if flg&flgPreRelease != 0 && !v.PreRelease {
			log.Debug("ignore non-prerelease: %d", v.ID)
			return false
		}
	}

	if o.Branch != "" && o.Branch != v.TargetCommitish {
		log.Debug("ignore release that branch does not matched to %q: %d", o.Branch, v.ID)
		return false
	}
	return true
}

var errEOL = errors.New("eol")

func listup(ghc *github.Client, flg int, o *Option) ([]*github.Release, error) {
	list := []*github.Release{}
	nitem := uint64(0)

	if err := ghc.FetchRelease(1, o.ItemsPerPage, func(v *github.Release, _ int) error {
		if !isListTarget(v, flg, o) {
			return nil
		} else if o.BranchExists {
			if b, err := ghc.GetBranch(v.TargetCommitish); err != nil {
				return err
			} else if b == nil {
				// branch does not exists
				log.Debug("ignore release that branch %q does not exists: %d", v.TargetCommitish, v.ID)
				return nil
			}
		}

		list = append(list, v)
		nitem++
		if o.MaxItems > 0 && nitem >= o.MaxItems {
			return errEOL
		}

		return nil
	}); err != nil && !errors.Is(errEOL, err) {
		return nil, err
	}

	return list, nil
}

func AllReleases(ghc *github.Client, o *Option) ([]*github.Release, error) {
	return listup(ghc, flgDraftRelease|flgPreRelease, o)
}

func PreReleases(ghc *github.Client, o *Option) ([]*github.Release, error) {
	return listup(ghc, flgPreRelease, o)
}

func DraftReleases(ghc *github.Client, o *Option) ([]*github.Release, error) {
	return listup(ghc, flgDraftRelease, o)
}

func Releases(ghc *github.Client, o *Option) ([]*github.Release, error) {
	return listup(ghc, flgReleaseOnly, o)
}
