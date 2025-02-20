package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antiphp/fakegs"
	"github.com/antiphp/fakegs/agones"
	"github.com/hamba/cmd/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/urfave/cli/v2"
)

func run(c *cli.Context) error { //nolint:cyclop,funlen // Readability.
	ctx, cancel := signal.NotifyContext(c.Context, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log, err := cmd.NewLogger(c)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	log.Info("Fake game server started")

	var client *agones.Client
	if c.IsSet(flagReadyAfter) || c.IsSet(flagAllocatedAfter) || c.IsSet(flagShutdownAfter) {
		sdkClient, err := agones.NewSDKClient(c.String(flagAgonesAddr))
		if err != nil {
			return fmt.Errorf("creating Agones SDK client: %w", err)
		}
		client = agones.NewClient(sdkClient)
		defer client.Close()

		go client.Run(ctx)
	}

	var exitErr exitHookError

	var optFns []fakegs.OptFunc
	if c.IsSet(flagExitAfter) {
		optFns = append(optFns, fakegs.WithExitAfter(c.Duration(flagExitAfter)))
	}
	if c.IsSet(flagExitCode) {
		exitErr.hookFn = func() { os.Exit(c.Int(flagExitCode)) }
	}
	if c.IsSet(flagExitSignal) {
		exitErr.hookFn = func() {
			_ = syscall.Kill(os.Getpid(), syscall.Signal(c.Int(flagExitSignal)))

			time.Sleep(10 * time.Second) // Brace for impact.
			os.Exit(1)
		}
	}
	if c.IsSet(flagExitCode) && c.IsSet(flagExitSignal) {
		log.Warn("Multiple contradictory exit behaviors configured.")
	}
	if c.IsSet(flagReadyAfter) {
		optFns = append(optFns, fakegs.WithUpdateStateAfter(agones.StateReady, c.Duration(flagReadyAfter), client))
	}
	if c.IsSet(flagAllocatedAfter) {
		optFns = append(optFns, fakegs.WithUpdateStateAfter(agones.StateAllocated, c.Duration(flagAllocatedAfter), client))
	}
	if c.IsSet(flagShutdownAfter) {
		optFns = append(optFns, fakegs.WithUpdateStateAfter(agones.StateShutdown, c.Duration(flagShutdownAfter), client))
	}
	if !c.IsSet(flagExitAfter) {
		log.Warn("No exit condition configured; external signal required")
	}

	reason, err := fakegs.New(log, optFns...).Run(ctx)
	if err != nil {
		log.Error("Fake game server stopped with an error", lctx.Err(err))
		return exitErr.maybeExitCode(1)
	}

	log.Info("Fake game server stopped", lctx.Str("reason", reason))
	return exitErr.maybeExitCode(0)
}
