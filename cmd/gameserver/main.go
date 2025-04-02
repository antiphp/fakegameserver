// Package main runs the fake game server.
package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/antiphp/fakegameserver/internal/exiterror"
	"github.com/ettle/strcase"
	"github.com/hamba/cmd/v2"
	"github.com/urfave/cli/v2"
)

const (
	flagExitCode             = "exit-code"
	flagExitSignal           = "exit-signal"
	flagExitAfter            = "exit-after"
	flagAgonesDisabled       = "agones-disabled"
	flagAgonesAddr           = "agones-addr"
	flagReadyAfter           = "ready-after"
	flagAllocatedAfter       = "allocated-after"
	flagShutdownAfter        = "shutdown-after"
	flagExitOnShutdown       = "shutdown-causes-exit"
	flagHealthReportDelay    = "health-report-delay"
	flagHealthReportInterval = "health-report-interval"

	catExit   = "Exit behavior"
	catAgones = "Agones integration"
)

var version = "<unknown>"

var flags = cmd.Flags{
	&cli.IntFlag{
		Name:     flagExitCode,
		Usage:    "Exit with this code, when an exit condition is met.",
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagExitCode))},
		Category: catExit,
	},
	&cli.IntFlag{
		Name:     flagExitSignal,
		Usage:    "Send this signal, when an exit condition is met.",
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagExitSignal))},
		Category: catExit,
	},
	&cli.DurationFlag{
		Name:     flagExitAfter,
		Usage:    "Flag after which to exit.",
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagExitAfter))},
		Category: catExit,
	},
	&cli.BoolFlag{
		Name:     flagAgonesDisabled,
		Usage:    "Flag whether to disable the Agones integration.",
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagAgonesDisabled))},
		Category: catAgones,
	},
	&cli.StringFlag{
		Name:     flagAgonesAddr,
		Usage:    "Address to reach the Agones SDK server.",
		Value:    "localhost:9357",
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagAgonesAddr))},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagReadyAfter,
		Usage:    "Duration after which to transition to Agones state `Ready`.",
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagReadyAfter))},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name: flagAllocatedAfter,
		Usage: "Duration after which to transition to Agones state `Allocated`. The `Ready`, `Allocated` and `Shutdown` timers are stacked. The first " +
			"timer starts immediately.",
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagAllocatedAfter))},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name: flagShutdownAfter,
		Usage: "Duration after which to transition to Agones state `Shutdown`. The `Ready`, `Allocated` and `Shutdown` timers are stacked. The first " +
			"timer starts immediately.",
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagShutdownAfter))},
		Category: catAgones,
	},
	&cli.BoolFlag{
		Name: flagExitOnShutdown,
		Usage: "Intended to be used for local development, to compensate the lack of a SIGTERM that usually follows a `Shutdown` in Agones cluster " +
			"environment.",
		EnvVars:     []string{strcase.ToSNAKE(prefixEnv(flagExitOnShutdown))},
		DefaultText: "'auto' - which enables the flag only if Agones runs in local development mode",
		Category:    catAgones,
	},
	&cli.DurationFlag{
		Name:     flagHealthReportDelay,
		Usage:    "Period after which the first Agones health report is sent.",
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagHealthReportDelay))},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagHealthReportInterval,
		Usage:    "Interval for the Agones health report.",
		Value:    5 * time.Second,
		EnvVars:  []string{strcase.ToSNAKE(prefixEnv(flagHealthReportInterval))},
		Category: catAgones,
	},
}.Merge(cmd.MonitoringFlags)

func main() {
	app := cli.NewApp()
	app.Name = "Fake Game Server with Agones integration"
	app.Version = version
	app.Flags = flags
	app.Action = run

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		var exitErr *exiterror.ExitError
		if errors.As(err, &exitErr) {
			exitErr.RunHooks()
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func prefixEnv(flag string) string {
	return "FAKEGAMESERVER_" + flag
}
