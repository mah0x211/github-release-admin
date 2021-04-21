package download

import (
	"strconv"

	"github-release-admin/getopt"
	"github-release-admin/github"
	"github-release-admin/log"
)

type ReleaseOption struct {
	Option
	ReleaseID int64
}

func (o *ReleaseOption) SetArg(arg string) bool {
	if o.ReleaseID == 0 {
		// verify release-id
		v, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			log.Error("invalid <release-id> argument %w", err)
			Usage(1)
		} else if v <= 0 {
			log.Error("<release-id> must be greater than 0")
			Usage(1)
		}
		o.ReleaseID = v
		return true
	}

	return o.Option.SetArg(arg)
}

func handleDownloadRelease(c *github.Client, args []string) {
	o := &ReleaseOption{}
	getopt.Parse(o, args)
	if o.ReleaseID == 0 || o.Name == "" {
		log.Error("invalid arguments")
		Usage(1)
	} else if v, err := c.GetRelease(int(o.ReleaseID)); err != nil {
		log.Fatalf("failed to get release: %v", err)
	} else if v == nil || !download(c, v.Assets, o.Name, o.NoDryRun) {
		log.Fatal("release not found")
	}
}
