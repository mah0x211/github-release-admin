package download

import (
	"encoding/json"
	"os"
	"strings"

	"github-release-admin/github"
	"github-release-admin/log"
)

var osExit = os.Exit

func Usage(exitcode int) {
	println(`
Download release assets

Usage:
    download <release-id> <filename> [--no-dry-run]
    download latest <filename> [--no-dry-run]
    download by-tag <tag>[@target] <filename> [--no-dry-run]

Arguments:
    <filename>             name of the asset to download.
    <release-id>           dowload from the specified release. (greater than 0)
    latest                 download from the lastest release.
    by-tag <tag>[@target]  download from the release associated with the
                           specified tag (and target).

Options:
    --no-dry-run           actually execute the request.
`)
	osExit(exitcode)
}

func isNotEmptyString(s string) bool {
	return strings.TrimSpace(s) != ""
}

type Option struct {
	Name     string
	NoDryRun bool
}

func (o *Option) SetArg(arg string) bool {
	if o.Name == "" {
		if isNotEmptyString(arg) {
			o.Name = arg
			return true
		}
	}
	log.Error("invalid argument")
	Usage(1)
	return false
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

func download(c *github.Client, v []github.Asset, name string, noDryRun bool) bool {
	for _, asset := range v {
		if asset.Name != name {
			continue
		}

		b, err := json.MarshalIndent(asset, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("download asset %d: %s", asset.ID, b)
		if noDryRun {
			if err = c.DownloadAsset(asset.ID, asset.Name); err != nil {
				log.Fatalf("error: %v", err)
			}
		}
		return true
	}
	return false
}

func Run(c *github.Client, args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	switch arg {
	case "latest":
		handleDownloadLatest(c, args[1:])

	case "by-tag":
		handleDownloadByTagName(c, args[1:])

	default:
		handleDownloadRelease(c, args)
	}
}
