# Fake Game Server with Agones Integration

This fake game server is a standalone executable with the goal of acting like a game server within a container environment.

The configurable features are:

- Transition into another Agones state
- Exit after a while
- Exit with a pre-defined code
- Crash with a pre-defined signal.

Potential features for the future are:

- Ignore SIGTERM/SIGINT
- Specified jitter for exit timers to make load tests more realistic
- Configurable memory consumption
- Configurable CPU consumption

## Usage

```
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

   --agones-addr value      Agones address. (default: "localhost:9357") [$AGONES_ADDR]
   --allocated-after value  Set Agones state to Allocated after the elapsed duration. State durations, if set, are stacked in order (Ready -> Allocated -> Shutdown). (default: 0s) [$ALLOCATED_AFTER]
   --ready-after value      Set Agones state to Ready after the elapsed duration. State durations, if set, are stacked in order (Ready -> Allocated -> Shutdown). (default: 0s) [$READY_AFTER]
   --shutdown-after value   Set Agones state to Shutdown after the elapsed duration. State durations, if set, are stacked in order (Ready -> Allocated -> Shutdown). (default: 0s) [$SHUTDOWN_AFTER]

   Exit behavior

   --exit-after value   Exit after the elapsed duration. (default: 0s) [$EXIT_AFTER]
   --exit-code value    Exit with this code. (default: 0) [$EXIT_CODE]
   --exit-signal value  Exit with this signal. (default: 0) [$EXIT_SIGNAL]

   Logging

   --log.ctx value [ --log.ctx value ]  A list of context field appended to every log. Format: key=value. [$LOG_CTX]
   --log.format value                   Specify the format of logs. Supported formats: 'logfmt', 'json', 'console'. [$LOG_FORMAT]
   --log.level value                    Specify the log level. e.g. 'trace', 'debug', 'info', 'error'. (default: "info") [$LOG_LEVEL]
```
