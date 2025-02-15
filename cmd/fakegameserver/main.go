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
)

var (
	serviceName = "fakegameserver"
	version     = "<unknown>"
)

var flags = cmd.Flags{
	&cli.IntFlag{
		Name:    flagExitCode,
		Usage:   "ExitBehavior with this code.",
		EnvVars: []string{strcase.ToSNAKE(flagExitCode)},
	},
	&cli.IntFlag{
		Name:    flagExitSignal,
		Usage:   "ExitBehavior with this signal.",
		EnvVars: []string{strcase.ToSNAKE(flagExitSignal)},
	},
	&cli.DurationFlag{
		Name:    flagExitAfter,
		Usage:   "ExitBehavior after the elapsed duration.",
		EnvVars: []string{strcase.ToSNAKE(flagExitAfter)},
	},
	&cli.StringFlag{
		Name:    flagAgonesAddr,
		Usage:   "Agones address.",
		Value:   "localhost:9357",
		EnvVars: []string{strcase.ToSNAKE(flagAgonesAddr)},
	},
	&cli.DurationFlag{
		Name:    flagReadyAfter,
		Usage:   "Set Agones status to Ready after the elapsed duration. State durations are stacked in order (Ready -> Allocated -> Shutdown).",
		EnvVars: []string{strcase.ToSNAKE(flagReadyAfter)},
	},
	&cli.DurationFlag{
		Name:    flagAllocatedAfter,
		Usage:   "Set Agones status to Allocated after the elapsed duration. State durations are stacked in order (Ready -> Allocated -> Shutdown).",
		EnvVars: []string{strcase.ToSNAKE(flagAllocatedAfter)},
	},
	&cli.DurationFlag{
		Name:    flagShutdownAfter,
		Usage:   "Set Agones state to Shutdown after the elapsed duration. State durations are stacked in order (Ready -> Allocated -> Shutdown).",
		EnvVars: []string{strcase.ToSNAKE(flagShutdownAfter)},
	},
}.Merge(cmd.MonitoringFlags)

func main() {
	app := cli.NewApp()
	app.Name = "Fake Game Server for Agones"
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
