package main

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	fakegs "github.com/antiphp/fakegameserver"
	"github.com/hamba/cmd/v2/observe"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

func run(c *cli.Context) (err error) {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	obsvr, err := observe.NewFromCLI(c, serviceName, &observe.Options{
		LogTimestamps: true,
		LogTimeFormat: logger.TimeFormatISO8601,
		TracingAttrs:  []attribute.KeyValue{semconv.ServiceVersionKey.String(version)},
	})
	if err != nil {
		return fmt.Errorf("creating observer: %w", err)
	}
	log := obsvr.Log

	log.Info("Fake game server started")

	var agones *fakegs.Agones
	if c.IsSet(flagReadyAfter) || c.IsSet(flagAllocatedAfter) || c.IsSet(flagShutdownAfter) {
		client, err := fakegs.NewAgonesClient(c.String(flagAgonesAddr))
		if err != nil {
			return fmt.Errorf("creating Agones client: %w", err)
		}
		agones = fakegs.NewAgones(client)
		defer agones.Close()
		defer cancel() // Agones is configured to re-try forever unless the context gets cancelled.

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
		log.Warn("Both exit hooks configured")
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
		return exitErr.appendExitCode(1)
	}

	log.Info("Fake game server stopped", lctx.Str("reason", reason))
	return exitErr.appendExitCode(0)
}
