package main

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/antiphp/fakegs"
	"github.com/hamba/cmd/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/urfave/cli/v2"
)

func run(c *cli.Context) error { //nolint:cyclop,funlen // Readability.
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	log, err := cmd.NewLogger(c)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	log.Info("Fake game server started")

	var agones *fakegs.Agones
	if c.IsSet(flagReadyAfter) || c.IsSet(flagAllocatedAfter) || c.IsSet(flagShutdownAfter) {
		client, err := fakegs.NewAgonesClient(c.String(flagAgonesAddr))
		if err != nil {
			return fmt.Errorf("creating Agones client: %w", err)
		}
		agones = fakegs.NewAgones(client)
		defer agones.Close()

		go agones.Run(ctx)
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
		optFns = append(optFns, fakegs.WithUpdateStateAfter(fakegs.StateReady, c.Duration(flagReadyAfter), agones))
	}
	if c.IsSet(flagAllocatedAfter) {
		optFns = append(optFns, fakegs.WithUpdateStateAfter(fakegs.StateAllocated, c.Duration(flagAllocatedAfter), agones))
	}
	if c.IsSet(flagShutdownAfter) {
		optFns = append(optFns, fakegs.WithUpdateStateAfter(fakegs.StateShutdown, c.Duration(flagShutdownAfter), agones))
	}
	if !c.IsSet(flagExitAfter) {
		log.Warn("No exit condition configured")
	}

	reason, err := fakegs.New(log, optFns...).Run(ctx)
	if err != nil {
		log.Error("Fake game server stopped with an error", lctx.Err(err))
		return exitErr.maybeExitCode(1)
	}

	log.Info("Fake game server stopped", lctx.Str("reason", reason))
	return exitErr.maybeExitCode(0)
}
