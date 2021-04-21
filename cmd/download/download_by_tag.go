package download

import (
	"strings"

	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
)

type TagOption struct {
	Option
	TagName         string
	TargetCommitish string
}

func (o *TagOption) SetArg(arg string) bool {
	if o.TagName == "" {
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
		log.Error("invalid arguments")
		Usage(1)
	}

	return o.Option.SetArg(arg)
}

func handleDownloadByTagName(c *github.Client, args []string) {
	o := &TagOption{}
	getopt.Parse(o, args)
	if o.TagName == "" || o.Name == "" {
		log.Error("invalid arguments")
		Usage(1)
	} else if v, err := c.GetReleaseByTagName(o.TagName); err != nil {
		log.Fatalf("failed to get release: %v", err)
	} else if v == nil ||
		(o.TargetCommitish != "" && v.TargetCommitish != o.TargetCommitish) ||
		!download(c, v.Assets, o.Name, o.NoDryRun) {
		log.Fatal("release not found")
	}
}
