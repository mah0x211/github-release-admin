package download

import (
	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
)

func handleDownloadLatest(c *github.Client, args []string) {
	o := &Option{}
	getopt.Parse(o, args)
	if o.Name == "" {
		log.Error("invalid arguments")
		Usage(1)
	} else if v, err := c.GetReleaseLatest(); err != nil {
		log.Fatalf("failed to get release: %v", err)
	} else if v == nil || !download(c, v.Assets, o.Name, o.NoDryRun) {
		log.Fatal("release not found")
	}
}
