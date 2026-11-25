# Signal

Send a signal to a process.

Sends TERM, KILL, HUP, INT, USR1 or USR2 to one process with kill, and reports
changed: true. A process that the host does not have is not found. A process ID
below 1 is refused, because kill sends to a group of processes for 0 and for a
negative ID.

```bash
osapi client node process signal [flags]
```

## Examples

```bash
# Stop a process
osapi client node process signal --target web-01 --pid 1234 --signal TERM
```

## Result

Stop a process, on one host:

```ansi
$ osapi client node process signal --target web-01 --pid 1234 --signal TERM
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mPID   [0m[38;2;10;143;176mSIGNAL[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  1234  TERM

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node process signal --target _all --pid 1234 --signal TERM
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mPID   [0m[38;2;10;143;176mSIGNAL[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  1234  TERM
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  1234  TERM

  3 hosts: [38;2;103;234;147m2 changed[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": true,
      "hostname": "web-01",
      "pid": 1234,
      "signal": "TERM",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example        | On the host            |
| -------------- | ---------------------- |
| Stop a process | runs `kill -TERM 1234` |

The command fails in these cases:

| Call                                  | Flags                       | Fails with |
| ------------------------------------- | --------------------------- | ---------- |
| A process that the host does not have | `--pid 99999 --signal TERM` | not found  |
| A process ID below 1                  | `--pid -1 --signal TERM`    | min        |
| A signal that osapi does not send     | `--pid 1234 --signal STOP`  | enum       |

## Permission

It needs the `process:execute` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                                                  | Default |
| -------------- | -------------------------------------------------------------------------------------------- | ------- |
| `--pid`        | The process ID, such as 1234 (required)                                                      |         |
| `--signal`     | The signal: TERM, KILL, HUP, INT, USR1 or USR2 (TERM, KILL, HUP, INT, USR1, USR2) (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                                       | `_all`  |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
