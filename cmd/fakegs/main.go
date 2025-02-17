// Package main runs the fake game server.
package main

import (
	"context"
	"errors"
	"os"

	"github.com/ettle/strcase"
	"github.com/hamba/cmd/v2"
	"github.com/urfave/cli/v2"
)

const (
	flagExitCode       = "exit-code"
	flagExitSignal     = "exit-signal"
	flagExitAfter      = "exit-after"
	flagAgonesAddr     = "agones-addr"
	flagReadyAfter     = "ready-after"
	flagAllocatedAfter = "allocated-after"
	flagShutdownAfter  = "shutdown-after"

	catExit   = "Exit behavior"
	catAgones = "Agones integration"
)

var version = "<unknown>"

var flags = cmd.Flags{
	&cli.IntFlag{
		Name:     flagExitCode,
		Usage:    "Exit with this code.",
		EnvVars:  []string{strcase.ToSNAKE(flagExitCode)},
		Category: catExit,
	},
	&cli.IntFlag{
		Name:     flagExitSignal,
		Usage:    "Exit with this signal.",
		EnvVars:  []string{strcase.ToSNAKE(flagExitSignal)},
		Category: catExit,
	},
	&cli.DurationFlag{
		Name:     flagExitAfter,
		Usage:    "Exit after the elapsed duration.",
		EnvVars:  []string{strcase.ToSNAKE(flagExitAfter)},
		Category: catExit,
	},
	&cli.StringFlag{
		Name:     flagAgonesAddr,
		Usage:    "Agones address.",
		Value:    "localhost:9357",
		EnvVars:  []string{strcase.ToSNAKE(flagAgonesAddr)},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagReadyAfter,
		Usage:    "Set Agones state to Ready after the elapsed duration. State durations, if set, are stacked in order (Ready -> Allocated -> Shutdown).",
		EnvVars:  []string{strcase.ToSNAKE(flagReadyAfter)},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagAllocatedAfter,
		Usage:    "Set Agones state to Allocated after the elapsed duration. State durations, if set, are stacked in order (Ready -> Allocated -> Shutdown).",
		EnvVars:  []string{strcase.ToSNAKE(flagAllocatedAfter)},
		Category: catAgones,
	},
	&cli.DurationFlag{
		Name:     flagShutdownAfter,
		Usage:    "Set Agones state to Shutdown after the elapsed duration. State durations, if set, are stacked in order (Ready -> Allocated -> Shutdown).",
		EnvVars:  []string{strcase.ToSNAKE(flagShutdownAfter)},
		Category: catAgones,
	},
}.Merge(cmd.LogFlags)

func main() {
	app := cli.NewApp()
	app.Name = "Fake Game Server with Agones integration"
	app.Version = version
	app.Flags = flags
	app.Action = run

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		var exitErr *exitHookError
		if errors.As(err, &exitErr) {
			exitErr.hookFn()
		}
		os.Exit(1)
	}
	os.Exit(0)
}
