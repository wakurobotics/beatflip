# beatflip - a minimal supervisor :fairy:

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

For a full list of commands run

```bash
beatflip --help
```

### beatflip supervise

Start the supervisor mode.

## Config-File

Configuration files are optional. You can also choose to either use command arguments or env vars. However, beatflip looks for an optional config file named `.beatflip.yml` in the current working directory. For a full example of the configuration file, please have a look at [`https://github.com/wakurobotics/beatflip/raw/master/.beatflip.yml`](https://github.com/wakurobotics/beatflip/raw/master/.beatflip.yml).

## OS Signals

Sending a SIGINT or SIGTERM to beatflip causes the supervisor to first terminate the supervised process by forwarding the respective signal, then terminating itself.

Sending a SIGHUB to the supervisor causes it to restart a currently running process.

```bash
beatflip supervise &
PID=$!
kill -HUP $PID  # --> supervised process will be restarted
```
