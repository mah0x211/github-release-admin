package download

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mah0x211/github-release-admin/github"
	"github.com/mah0x211/github-release-admin/log"
)

type Option struct {
	SaveAs string
	DryRun bool
}

func download(ghc *github.Client, v *github.Asset, o *Option) error {
	if log.Verbose {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to stringify the asset-info: %w", err)
		}
		log.Debug("download asset %d: %s", v.ID, b)
	}

	if o.DryRun {
		return nil
	}

	saveAs := v.Name
	if o.SaveAs = strings.TrimSpace(o.SaveAs); o.SaveAs != "" {
		saveAs = o.SaveAs
	}

	return ghc.DownloadAsset(v.ID, saveAs)
}

func selectAsset(v []github.Asset, name string) *github.Asset {
	for _, asset := range v {
		if asset.Name == name {
			return &asset
		}
	}
	return nil
}

var ErrNotFound = fmt.Errorf("not found")

func Latest(ghc *github.Client, name string, o *Option) error {
	var a *github.Asset

	if v, err := ghc.GetReleaseLatest(); err != nil {
		return err
	} else if v == nil {
		return ErrNotFound
	} else if a = selectAsset(v.Assets, name); a == nil {
		return ErrNotFound
	}

	return download(ghc, a, o)
}

func ByTagName(ghc *github.Client, tag, targetCommitish, name string, o *Option) error {
	var a *github.Asset

	if v, err := ghc.GetReleaseByTagName(tag); err != nil {
		return err
	} else if v == nil {
		return ErrNotFound
	} else if targetCommitish != "" && v.TargetCommitish != targetCommitish {
		return ErrNotFound
	} else if a = selectAsset(v.Assets, name); a == nil {
		return ErrNotFound
	}

	return download(ghc, a, o)
}

func Release(ghc *github.Client, id int, name string, o *Option) error {
	var a *github.Asset

	if v, err := ghc.GetRelease(id); err != nil {
		return err
	} else if v == nil {
		return ErrNotFound
	} else if a = selectAsset(v.Assets, name); a == nil {
		return ErrNotFound
	}

	return download(ghc, a, o)
}
