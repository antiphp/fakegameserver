// Package main runs the fake game server.
package main

import (
	"context"
	"errors"
	"os"
	"time"

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
		EnvVars:  []string{strcase.ToSNAKE(flagExitCode)},
		Category: catExit,
	},
	&cli.IntFlag{
		Name:     flagExitSignal,
		Usage:    "Send this signal, when an exit condition is met.",
		EnvVars:  []string{strcase.ToSNAKE(flagExitSignal)},
		Category: catExit,
	},
	&cli.DurationFlag{
		Name:     flagExitAfter,
		Usage:    "after after which to exit.",
		EnvVars:  []string{strcase.ToSNAKE(flagExitAfter)},
		Category: catExit,
	},
	&cli.BoolFlag{
		Name:     flagAgonesDisabled,
		Usage:    "Flag whether to disable the Agones integration.",
		EnvVars:  []string{strcase.ToSNAKE(flagAgonesDisabled)},
		Category: catAgones,
	},
	&cli.StringFlag{
		Name:     flagAgonesAddr,
		Usage:    "Address to reach the Agones Client server.",
		Value:    "localhost:9357",
		EnvVars:  []string{strcase.ToSNAKE(flagAgonesAddr)},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagReadyAfter,
		Usage:    "Duration after which to transition to the Agones state Ready.",
		EnvVars:  []string{strcase.ToSNAKE(flagReadyAfter)},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagAllocatedAfter,
		Usage:    "Duration after which to transition to the Agones state Allocated. If a Ready timer is set, the timer starts after the Ready transition.",
		EnvVars:  []string{strcase.ToSNAKE(flagAllocatedAfter)},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagShutdownAfter,
		Usage:    "Duration after which to transition to the Agones state Shutdown. If a Ready or Allocated timer is set, the timer starts after the Ready or Allocated transition.",
		EnvVars:  []string{strcase.ToSNAKE(flagShutdownAfter)},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagHealthReportDelay,
		Usage:    "Period after which the first Agones health report is sent.",
		EnvVars:  []string{strcase.ToSNAKE(flagHealthReportDelay)},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagHealthReportInterval,
		Usage:    "Interval for the Agones health report.",
		Value:    5 * time.Second,
		EnvVars:  []string{strcase.ToSNAKE(flagHealthReportInterval)},
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
		var exitErr *exitError
		if errors.As(err, &exitErr) {
			exitErr.runHooks()
		}
		os.Exit(1)
	}
	os.Exit(0)
}
