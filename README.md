# Fake Game Server with Agones Integration

The Fake Game Server, or just fakegs, is a standalone executable with no game-related functionality but an Agones (https://agones.dev/) integration.
It is intended for testing game server orchestration without the need for a real game server executable.

## Features

Supported features are:

- Transition into [Agones](https://agones.dev/) the states `Ready`, `Allocated` and `Shutdown` after a configurable duration,
- Exit after a configured duration,
- Exit with a configured exit code,
- Exit with a configured signal.

### Agones

The Agones integration allows scheduled state transitions.
The state transitions are performed one after another, if set, in the order `Ready`, `Allocated`, `Shutdown`.

| Argument             | Environment                       | Type     | Default          | Example | Description                                                                                                                                                     |
|----------------------|-----------------------------------|----------|------------------|---------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--agones-addr`      | `FAKEGAMESERVER_AGONES_ADDR`      | `string` | `localhost:9357` | -       | Address to reach the Agones SDK server.                                                                                                                         |
| `--ready-after`      | `FAKEGAMESERVER_READY_AFTER`      | `string` | `0s` (disabled)  | `10s`   | Duration after which to transition to Agones state `Ready`.                                                                                                     |
| `--allocated-after`  | `FAKEGAMESERVER_ALLOCATED_AFTER`  | `string` | `0s` (disabled)  | `5s`    | Duration after which to transition to Agones state `Allocated`. The `Ready`, `Allocated` and `Shutdown` timers are stacked. The first timer starts immediately. |
| `--shutdown-after`   | `FAKEGAMESERVER_SHUTDOWN_AFTER`   | `string` | `0s` (disabled)  | `30s`   | Duration after which to transition to Agones state `Shutdown`. The `Ready`, `Allocated` and `Shutdown` timers are stacked. The first timer starts immediately.  |
| `--exit-on-shutdown` | `FAKEGAMESERVER_EXIT_ON_SHUTDOWN` | `bool`   | (auto)           | `true`  | Intended to be used for local development, to compensate the lack of a SIGTERM that usually follows a `Shutdown` in Agones cluster environment.                 |

With the given example values, the fakegs transitions to state `Ready` after `10s`, then `5s` later to `Allocated` (in total after `15s`),
and `30s` later to `Shutdown` (in total after `45s`), and then exits.

### Exit Behavior

| Argument        | Environment                  | Type     | Default         | Example          | Description                                         |
|-----------------|------------------------------|----------|-----------------|------------------|-----------------------------------------------------|
| `--exit-after`  | `FAKEGAMESERVER_EXIT_AFTER`  | `string` | `0s` (disabled) | `2m`             | Duration after which to exit.                       |
| `--exit-code`   | `FAKEGAMESERVER_EXIT_CODE`   | `int`    | (auto)          | `0`              | Exit with this code, when an exit condition is met. |
| `--exit-signal` | `FAKEGAMESERVER_EXIT_SIGNAL` | `int`    | (none)          | `11` (`SIGSEGV`) | Send this signal, when an exit condition is met.    |

With the given example values, the fakegs exits after `2m` with a crash (`SIGSEGV`) (`--exit-signal` would overwrite `--exit-code` as the exit condition).

## Usage

```$ go run ./cmd/fakegs/ --help
NAME:
   Fake Game Server with Agones integration - A new cli application

USAGE:
   Fake Game Server with Agones integration [global options] command [command options]

VERSION:
   <unknown>

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

   Agones integration

   --agones-addr value              Address to reach the Agones SDK server. (default: "localhost:9357") [$FAKEGAMESERVER_AGONES_ADDR]
   --agones-disabled                Flag whether to disable the Agones integration. (default: false) [$FAKEGAMESERVER_AGONES_DISABLED]
   --allocated-after Allocated      Duration after which to transition to Agones state Allocated. The `Ready`, `Allocated` and `Shutdown` timers are stacked. The first timer starts immediately. (default: 0s) [$FAKEGAMESERVER_ALLOCATED_AFTER]
   --health-report-delay value      Period after which the first Agones health report is sent. (default: 0s) [$FAKEGAMESERVER_HEALTH_REPORT_DELAY]
   --health-report-interval value   Interval for the Agones health report. (default: 5s) [$FAKEGAMESERVER_HEALTH_REPORT_INTERVAL]
   --ready-after Ready              Duration after which to transition to Agones state Ready. (default: 0s) [$FAKEGAMESERVER_READY_AFTER]
   --shutdown-after Shutdown        Duration after which to transition to Agones state Shutdown. The `Ready`, `Allocated` and `Shutdown` timers are stacked. The first timer starts immediately. (default: 0s) [$FAKEGAMESERVER_SHUTDOWN_AFTER]
   --shutdown-causes-exit Shutdown  Meant to be used for local development, to compensate the lack of a SIGTERM that usually follows a Shutdown in Agones cluster environment. (default: 'auto' - which enables the flag only if Agones runs in local development mode) [$FAKEGAMESERVER_SHUTDOWN_CAUSES_EXIT]

   Exit behavior

   --exit-after value   Flag after which to exit. (default: 0s) [$FAKEGAMESERVER_EXIT_AFTER]
   --exit-code value    Exit with this code, when an exit condition is met. (default: 0) [$FAKEGAMESERVER_EXIT_CODE]
   --exit-signal value  Send this signal, when an exit condition is met. (default: 0) [$FAKEGAMESERVER_EXIT_SIGNAL]

   Logging

   --log.ctx value [ --log.ctx value ]  A list of context field appended to every log. Format: key=value. [$LOG_CTX]
   --log.format value                   Specify the format of logs. Supported formats: 'logfmt', 'json', 'console'. [$LOG_FORMAT]
   --log.level value                    Specify the log level. e.g. 'trace', 'debug', 'info', 'error'. (default: "info") [$LOG_LEVEL]

   Profiling

   --profiling.dsn value                                The address to the Pyroscope server, in the format: 'http://basic:auth@server:port?token=auth-token&tenantid=tenant-id'. [$PROFILING_DSN]
   --profiling.tags value [ --profiling.tags value ]    A list of tags appended to every profile. Format: key=value. [$PROFILING_TAGS]
   --profiling.types value [ --profiling.types value ]  The type of profiles to include. Defaults to all. [$PROFILING_TYPES]
   --profiling.upload-rate value                        The rate at which profiles are uploaded. (default: 15s) [$PROFILING_UPLOAD_RATE]

   Stats

   --stats.dsn value                          The DSN of a stats backend. [$STATS_DSN]
   --stats.interval value                     The frequency at which the stats are reported. (default: 1s) [$STATS_INTERVAL]
   --stats.prefix value                       The prefix of the measurements names. [$STATS_PREFIX]
   --stats.tags value [ --stats.tags value ]  A list of tags appended to every measurement. Format: key=value. [$STATS_TAGS]

   Tracing

   --tracing.endpoint value                             The tracing backend endpoint. [$TRACING_ENDPOINT]
   --tracing.endpoint-insecure                          Determines if the endpoint is insecure. (default: false) [$TRACING_ENDPOINT_INSECURE]
   --tracing.exporter value                             The tracing backend. Supported: 'zipkin', 'otlphttp', 'otlpgrpc'. [$TRACING_EXPORTER]
   --tracing.headers value [ --tracing.headers value ]  A list of headers appended to every trace when supported by the exporter. Format: key=value. [$TRACING_HEADERS]
   --tracing.ratio value                                The ratio between 0 and 1 of sample traces to take. (default: 0.5) [$TRACING_RATIO]
   --tracing.tags value [ --tracing.tags value ]        A list of tags appended to every trace. Format: key=value. [$TRACING_TAGS]
```

## Example

```shell
$ go run ./fakegameserver/cmd/gameserver/ --allocated-after=30s --shutdown-after=10s
ts=2025-03-21T22:40:33.888+01:00 lvl=info msg="Game server started" svc=fakegameserver
ts=2025-03-21T22:40:34.388+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Agones connecting" type=agonesConnection error="rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9357: connect: connection refused\""
ts=2025-03-21T22:40:37.889+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Agones connection established" type=agonesConnection
ts=2025-03-21T22:40:37.889+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Game server became healthy" type=healthStatus
ts=2025-03-21T22:40:37.889+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Health reported" type=agonesReportHealth
ts=2025-03-21T22:40:37.889+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Agones state change received for Ready" type=agonesUpdate
ts=2025-03-21T22:40:43.789+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Health reported" type=agonesReportHealth
ts=2025-03-21T22:40:49.889+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Health reported" type=agonesReportHealth
ts=2025-03-21T22:40:55.888+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Health reported" type=agonesReportHealth
ts=2025-03-21T22:41:00.789+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Health reported" type=agonesReportHealth
ts=2025-03-21T22:41:06.889+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Health reported" type=agonesReportHealth
ts=2025-03-21T22:41:07.888+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Requesting Agones state update to Allocated" type=agonesRequestUpdate
ts=2025-03-21T22:41:07.888+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Agones state updated" type=agonesUpdate
ts=2025-03-21T22:41:07.888+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Agones state change received for Allocated" type=agonesUpdate
ts=2025-03-21T22:41:12.888+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Health reported" type=agonesReportHealth
ts=2025-03-21T22:41:17.888+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Requesting Agones state update to Shutdown" type=agonesRequestUpdate
ts=2025-03-21T22:41:17.888+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Agones state updated" type=agonesUpdate
ts=2025-03-21T22:41:17.888+01:00 lvl=info msg="Game server message received" svc=fakegameserver desc="Agones state changed to Shutdown, emulating the behavior of Agones in a non-local development environment with SIGTERM" type=exit error="signal 15"
ts=2025-03-21T22:41:17.888+01:00 lvl=info msg="Game server stopped with error" svc=fakegameserver exit="signal 15"
signal: terminated
```

It is required to have an Agones SDK server running under `localhost:9357`, either in a separate Kubernetes container, or locally from
within https://github.com/googleforgames/agones with `go run ./cmd/sdk-server --local`.

## Docker

A container image is available under `docker.io/antiphp/fakegameserver`.
