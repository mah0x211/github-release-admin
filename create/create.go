package create

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mah0x211/github-release-admin/github"
	"github.com/mah0x211/github-release-admin/log"
	"github.com/mah0x211/github-release-admin/readdir"
)

type Option struct {
	TagName         string
	TargetCommitish string
	Filename        string
	Title           string
	Body            string
	Dirname         string
	AsRegex         bool
	AsPosix         bool
	Draft           bool
	PreRelease      bool
	DryRun          bool
}

func upload(ghc *github.Client, v *github.Release, pathname string, o *Option) error {
	f, err := os.Open(pathname)
	if err != nil {
		return err
	}
	defer f.Close()

	// detect content-length and content-type
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	} else if _, err = f.Seek(0, 0); err != nil {
		return err
	}
	size := int64(len(b))
	mime := http.DetectContentType(b)
	name := filepath.Base(pathname)

	log.Debug("upload %s %d byte (%s)", name, size, mime)
	if !o.DryRun {
		return v.UploadAsset(ghc, name, f, size, mime)
	}
	return nil
}

func Release(ghc *github.Client, o *Option, r *readdir.Reader, assets []string) error {
	if len(assets) == 0 {
		return nil
	}

	var v *github.Release
	var err error
	if o.DryRun {
		v = &github.Release{}
	} else if v, err = ghc.CreateRelease(
		o.TagName, o.TargetCommitish, o.Title, o.Body, o.Draft, o.PreRelease,
	); err != nil {
		return err
	}

	if log.Verbose {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		log.Debug("create release %s", b)
	}

	// upload asset files
	for _, pathname := range assets {
		if err = upload(ghc, v, pathname, o); err != nil {
			if !o.DryRun {
				if err := ghc.DeleteRelease(v.ID); err != nil {
					log.Errorf("failed to delete the failed release: %v", err)
				}
			}
			return err
		}
	}

	return nil
}
