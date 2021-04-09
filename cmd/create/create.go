package create

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
	"github-release-admin/readdir"
)

var osExit = os.Exit

func Usage(code int) {
	println(`
Create release and upload asset files.

Usage:
    create <tag>[@<target>] <filename>
           [--title=<title>] [--body=<body>]
           [--dir=<path/to/dir>] [--regex] [--posix]
           [--no-draft] [--no-prerelease] [--no-dry-run]

Arguments:
    tag                 specify an existing tag, or create a new tag.
                        (e.g. v1.0.0)
    target              specify a branch, or commish. (e.g. master)
    filename            name of the asset file to upload. (e.g. myasset.tar.gz)

Options:
    --title=<title>     release title.
    --body=<body>       describe this release.
    --dir=<path/to/dir> reads the file from this directory.
    --regex             compile filename as regular expressions.
    --posix             compile filename as POSIX ERE (egrep).
    --no-draft          save as non-draft release.
    --no-prerelease     save as non-prerelease (production ready).
    --no-dry-run        actually execute the request.
`)
	osExit(1)
}

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

func upload(c *github.Client, o *Option, id int, pathname string) error {
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

	log.Printf("upload %s %d byte (%s)", name, size, mime)
	if !o.DryRun {
		return c.UploadAsset(id, name, f, size, mime)
	}
	return nil
}

func handleRelease(c *github.Client, o *Option, r *readdir.Reader) {
	assets, err := r.Read()
	if err != nil {
		log.Fatalf("failed to readdir(): %v", err)
	} else if len(assets) == 0 {
		log.Print("asset files not found")
		return
	}

	log.Printf(`create release:
  tag        : %q
  target     : %q
  title      : %q
  body       : %q
  draft      : %t
  pre-release: %t`,
		o.TagName, o.TargetCommitish, o.Title, o.Body,
		o.Draft, o.PreRelease)

	var rel *github.Release
	if o.DryRun {
		rel = &github.Release{}
	} else if rel, err = c.CreateRelease(
		o.TagName, o.TargetCommitish, o.Title, o.Body, o.Draft, o.PreRelease,
	); err != nil {
		log.Fatalf("failed to create release: %v", err)
	}

	// upload asset files
	for _, pathname := range assets {
		if err = upload(c, o, rel.ID, pathname); err != nil {
			log.Errorf("failed to upload %q: %v", pathname, err)
			if !o.DryRun {
				if err = c.DeleteRelease(rel.ID); err != nil {
					log.Fatalf("failed to delete release: %v", err)
				}
			}
			break
		}
	}

	if o.DryRun {
		if err = c.DeleteRelease(rel.ID); err != nil {
			log.Fatalf("failed to delete release: %v", err)
		}
	}
}

func Run(c *github.Client, args []string) {
	o := &Option{
		Draft:      true,
		PreRelease: true,
		DryRun:     true,
	}
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
		log.Errorf("filename cannot be compiled as regular expressions: %v", err)
		Usage(1)
	}

	handleRelease(c, o, readdir.New(o.Dirname, o.Filename, re))
}
