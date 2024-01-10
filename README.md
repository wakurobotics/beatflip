[![tests](https://github.com/wakurobotics/beatflip/actions/workflows/tests.yml/badge.svg?branch=master)](https://github.com/wakurobotics/beatflip/actions/workflows/tests.yml) [![Windows Release](https://img.shields.io/badge/Windows-Download-blue)](https://github.com/wakurobotics/beatflip/releases/latest/download/beatflip-windows-amd64.exe) [![MacOS Release](https://img.shields.io/badge/MacOS-Download-blue)](https://github.com/wakurobotics/beatflip/releases/latest/download/beatflip-macos-amd64) [![Linux Release](https://img.shields.io/badge/Linux-Download-blue)](https://github.com/wakurobotics/beatflip/releases/latest/download/beatflip-linux-amd64) [![ARM Release](https://img.shields.io/badge/ARM-Download-blue)](https://github.com/wakurobotics/beatflip/releases/latest/download/beatflip-linux-arm)

# beatflip :fairy:

> a minimal supervisor

Beatflip is a minimal supervisor tailored to single-process applications. It comes with the following features:

- :green_heart: **Process uptime**: monitors processes and restarts them if exited
- :beverage_box: **OTA updates**: downloads newer versions against an update-server / S3 bucket, flips & restarts binaries
- :postbox: **Log handling**: forwards process output

## Installation

To install / update beatflip run

```bash
go install github.com/wakurobotics/beatflip@latest
```

## Usage

```bash
$ beatflip --help
Usage:
  beatflip [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        initialize a new configuration file in the current directoey
  supervise
  version

Flags:
      --config string   config file (default is '.beatflip.yml')
  -h, --help            help for beatflip

Use "beatflip [command] --help" for more information about a command.
```

### beatflip init

Initialize a new configuration file.

### beatflip supervise

Start the supervisor mode.

## Config-File

You can initialize a new configuration file via `beatflip init`.

By default, Beatflip looks for a config file named `.beatflip.yml` in the current working directory. This can be overridden via the `--config` flag.

For a full example of the configuration file, please have a look at [`https://github.com/wakurobotics/beatflip/raw/master/cmd/.beatflip.yml`](https://github.com/wakurobotics/beatflip/raw/master/cmd/.beatflip.yml).

## OS Signals

Sending a SIGINT or SIGTERM to beatflip causes the supervisor to first terminate the supervised process by forwarding the respective signal, then terminating itself.

Sending a SIGHUB to the supervisor causes it to restart a currently running process.

```bash
beatflip supervise &
PID=$!
kill -HUP $PID  # --> supervised process will be restarted
```
