package main

import (
	"fmt"
	"os/signal"
	"syscall"

	"github.com/antiphp/fakegameserver"
	"github.com/antiphp/fakegameserver/agones"
	"github.com/antiphp/fakegameserver/internal/exiterror"
	"github.com/hamba/cmd/v2/observe"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/urfave/cli/v2"
	"k8s.io/utils/ptr"
)

func run(c *cli.Context) error {
	ctx, cancel := signal.NotifyContext(c.Context, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	obsvr, err := observe.NewFromCLI(c, "fakegameserver", &observe.Options{
		LogTimeFormat: "2006-01-02T15:04:05.999Z07:00",
		LogTimestamps: true,
	})
	if err != nil {
		return fmt.Errorf("creating observer: %w", err)
	}

	obsvr.Log.Info("Game server started")

	gs := fakegameserver.New(obsvr.Log)
	healthStatus := fakegameserver.NewHealthStatus()
	healthStatus.Exclude(fakegameserver.MessageTypeInfo, fakegameserver.MessageTypeExit)

	if c.Duration(flagExitAfter) > 0 {
		gs.AddHandler(fakegameserver.NewExitTimer(c.Duration(flagExitAfter)))
	}
	if !c.Bool(flagAgonesDisabled) {
		sdkClient, err := agones.NewSDKClient(c.String(flagAgonesAddr))
		if err != nil {
			return fmt.Errorf("creating Agones sdk client: %w", err)
		}

		client := agones.NewClient(sdkClient)
		go client.Run(ctx)

		gs.AddHandler(fakegameserver.NewAgonesWatcher(client))
		healthStatus.WaitFor(fakegameserver.MessageTypeAgonesConnection)

		gs.AddHandler(fakegameserver.NewAgonesHealthReporter(client, c.Duration(flagHealthReportDelay), c.Duration(flagHealthReportInterval)))
		healthStatus.Exclude(fakegameserver.MessageTypeAgonesReportHealth)

		gs.AddHandler(fakegameserver.NewAgonesStateUpdater(client))

		stateTimer := fakegameserver.NewAgonesStateTimer()
		if c.IsSet(flagReadyAfter) {
			stateTimer.AddState(agones.StateReady, c.Duration(flagReadyAfter))
		}
		if c.IsSet(flagAllocatedAfter) {
			stateTimer.AddState(agones.StateAllocated, c.Duration(flagAllocatedAfter))
		}
		if c.IsSet(flagShutdownAfter) {
			stateTimer.AddState(agones.StateShutdown, c.Duration(flagShutdownAfter))
		}
		gs.AddHandler(stateTimer)

		gs.AddHandler(fakegameserver.NewAgonesShutdown(func() bool {
			return (client.IsLocal() && !c.IsSet(flagExitOnShutdown)) || c.Bool(flagExitOnShutdown)
		}))
	}

	gs.AddHandler(healthStatus)

	var code, sig *int
	if c.IsSet(flagExitCode) {
		code = ptr.To[int](c.Int(flagExitCode))
	}
	if c.IsSet(flagExitSignal) {
		sig = ptr.To[int](c.Int(flagExitSignal))
	}
	exitErr := exiterror.New(code, sig)

	reason, err := gs.Run(ctx)
	if err != nil {
		wrapErr := exiterror.Wrap(exitErr, err, exiterror.New(ptr.To(1), nil))
		obsvr.Log.Info("Game server stopped with error", lctx.Str("exit", wrapErr.Error()))

		return wrapErr
	}

	wrapErr := exiterror.Wrap(exitErr, exiterror.New(ptr.To(0), nil))
	obsvr.Log.Info("Game server stopped", lctx.Str("reason", reason), lctx.Str("exit", wrapErr.Error()))
	return wrapErr
}
