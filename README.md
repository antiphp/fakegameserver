# Fake Game Server with Agones Integration

The fake game server, or short fakegs, is a standalone executable that acts like a very simple game server in a container environment.
Its main goal is to be a drop-in replacement for systems managing game servers for testing purposes.

## Features

Supported features are:

- Transition into [Agones](https://agones.dev/) the states `Ready`, `Allocated` and `Shutdown` after a configurable duration,
- Exit after a configured duration,
- Exit with a pre-defined code,
- Crash with a pre-defined signal.

### Agones

The Agones integration allows scheduled state transitions.
The state transitions are performed one after another, if set, in the order `Ready`, `Allocated`, `Shutdown`.

| Argument            | Environment       | Type     | Default          | Example | Description                                                                                                                                    |
|---------------------|-------------------|----------|------------------|---------|------------------------------------------------------------------------------------------------------------------------------------------------|
| `--agones-addr`     | `AGONES_ADDR`     | `string` | `localhost:9357` | -       | Address to reach the Agones SDK server.                                                                                                        |
| `--ready-after`     | `READY_AFTER`     | `string` | `0s` (disabled)  | `10s`   | Duration after which to transition to Agones state `Ready`.                                                                                    |
| `--allocated-after` | `ALLOCATED_AFTER` | `string` | `0s` (disabled)  | `5s`    | Duration after which to transition to Agones state `Allocated`. When the ready timer is not set, the timer starts immediately.                 |
| `--shutdown-after`  | `SHUTDOWN_AFTER`  | `string` | `0s` (disabled)  | `30s`   | Duration after which to transition to Agones state `Shutdown`. When the ready and/or allocated timer is not set, the timer starts immediately. |

With the given example values, the fakegs transitions to state `Ready` after `10s`, then `5s` later to `Allocated` (in total after `15s`),
and `30s` later to `Shutdown` (in total after `45s`).

### Exit Behavior

| Argument        | Environment   | Type     | Default         | Example          | Description                                         |
|-----------------|---------------|----------|-----------------|------------------|-----------------------------------------------------|
| `--exit-after`  | `EXIT_AFTER`  | `string` | `0s` (disabled) | `2m`             | Duration after which to exit.                       |
| `--exit-code`   | `EXIT_CODE`   | `int`    | (auto)          | `0`              | Exit with this code, when an exit condition is met. |
| `--exit-signal` | `EXIT_SIGNAL` | `int`    | (none)          | `11` (`SIGSEGV`) | Send this signal, when an exit condition is met.    |

With the given example values, the fakegs exits after `2m` with a crash (`SIGSEGV`) (`--exit-signal` would overwrite `--exit-code` as the exit condition).

## Usage

```$ go run ./cmd/fakegs/ --help
NAME:
   Fake Game Server with Agones integration - A new cli application

USAGE:
   Fake Game Server with Agones integration [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

   Agones integration

   --agones-addr value          Address to reach the Agones SDK server. (default: "localhost:9357") [$AGONES_ADDR]
   --allocated-after Allocated  Duration after which to transition to Agones state Allocated. When the ready timer is not set, the timer starts immediately. (default: 0s) [$ALLOCATED_AFTER]
   --ready-after Ready          Duration after which to transition to Agones state Ready. (default: 0s) [$READY_AFTER]
   --shutdown-after Shutdown    Duration after which to transition to Agones state Shutdown. When the ready and/or allocated timer is not set, the timer starts immediately. (default: 0s) [$SHUTDOWN_AFTER]

   Exit behavior

   --exit-after value   Duration after which to exit. (default: 0s) [$EXIT_AFTER]
   --exit-code value    Exit with this code, when an exit condition is met. (default: 0) [$EXIT_CODE]
   --exit-signal value  Send this signal, when an exit condition is met. (default: 0) [$EXIT_SIGNAL]

   Logging

   --log.ctx value [ --log.ctx value ]  A list of context field appended to every log. Format: key=value. [$LOG_CTX]
   --log.format value                   Specify the format of logs. Supported formats: 'logfmt', 'json', 'console'. [$LOG_FORMAT]
   --log.level value                    Specify the log level. e.g. 'trace', 'debug', 'info', 'error'. (default: "info") [$LOG_LEVEL]
```

or with Docker: `docker.io/antiphp/fakegs`.
