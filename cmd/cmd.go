package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mah0x211/github-release-admin/github"
	"github.com/mah0x211/github-release-admin/log"
	"github.com/mah0x211/github-release-admin/util"
)

type StartFunc func(ctx context.Context, ghc *github.Client, args []string)
type UsageFunc func(code int)

func Start(startfn StartFunc, usagefn UsageFunc) int {
	ctx, cancel := context.WithCancel(context.Background())

	// setup signal receiver
	sigch := make(chan os.Signal, 1)
	signal.Ignore()
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer cancel()
		args := os.Args[1:]
		if len(args) > 0 && args[0] == "help" {
			usagefn(0)
		}

		// use GITHUB_REPOSITORY envvar as first argument
		if v, found := util.Getenv("GITHUB_REPOSITORY"); found && v != "" {
			args = append([]string{v}, args...)
		}
		if len(args) == 0 {
			usagefn(0)
		}

		ghc, err := github.New(ctx, args[0])
		if err != nil {
			log.Error(err)
			usagefn(1)
		}
		// use GITHUB_TOKEN if define
		if v, found := util.Getenv("GITHUB_TOKEN"); found {
			ghc.SetToken(v)
		}
		// use GITHUB_API_URL if define
		if v, found := util.Getenv("GITHUB_API_URL"); found {
			if err = ghc.SetURL(v); err != nil {
				log.Errorf("invalid GITHUB_API_URL environment variable: %v", err)
				usagefn(1)
			}
		}

		startfn(ctx, ghc, args[1:])
	}()

	select {
	case <-ctx.Done():
	case sig := <-sigch:
		log.Errorf("stop command by %s", sig)
		cancel()
		select {
		case <-ctx.Done():
		case sig = <-sigch:
			log.Debug("stop command immediately by %s", sig)
		}
		return int(sig.(syscall.Signal))
	}

	return util.ExitCode
}
