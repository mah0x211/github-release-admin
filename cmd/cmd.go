package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github-release-admin/github"
	"github-release-admin/log"
	"github-release-admin/util"
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
		if len(args) == 0 || args[0] == "help" {
			usagefn(0)
		}

		ghc, err := github.New(args[0])
		if err != nil {
			log.Error(err)
			usagefn(1)
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
