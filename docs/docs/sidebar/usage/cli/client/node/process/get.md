# Get

Get one process.

Reads one process by its ID, without the others. A process that the host does
not have is not found.

```bash
osapi client node process get [flags]
```

## Examples

```bash
# A process of a Linux host
osapi client node process get --target web-01 --pid 1234

# A process of a Mac
osapi client node process get --target web-01 --pid 1234
```

## Result

A process of a Linux host, on one host:

```ansi
$ osapi client node process get --target web-01 --pid 1234
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPID   [0m[38;2;10;143;176mNAME   [0m[38;2;10;143;176mUSER   [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mCPU PERCENT  [0m[38;2;10;143;176mMEM PERCENT  [0m[38;2;10;143;176mMEM RSS  [0m[38;2;10;143;176mCOMMAND    [0m[38;2;10;143;176mSTART TIME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      1234  [1msleep[0m  alice  sleep  5.00         1.00         20.0 MB  sleep 300  2026-01-02T02:55:45Z

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node process get --target _all --pid 1234
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPID   [0m[38;2;10;143;176mNAME   [0m[38;2;10;143;176mUSER   [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mCPU PERCENT  [0m[38;2;10;143;176mMEM PERCENT  [0m[38;2;10;143;176mMEM RSS  [0m[38;2;10;143;176mCOMMAND    [0m[38;2;10;143;176mSTART TIME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      1234  [1msleep[0m  alice  sleep  5.00         1.00         20.0 MB  sleep 300  2026-01-02T02:55:45Z
  [1mweb-02[0m    [38;2;103;234;147mok[0m      1234  [1msleep[0m  alice  sleep  5.00         1.00         20.0 MB  sleep 300  2026-01-02T02:55:45Z

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "command": "sleep 300",
      "cpu_percent": 5,
      "hostname": "web-01",
      "mem_percent": 1,
      "mem_rss": 20971520,
      "name": "sleep",
      "pid": 1234,
      "start_time": "2026-01-02T02:55:45Z",
      "state": "sleep",
      "status": "ok",
      "user": "alice"
    }
  ]
}
```

A process of a Mac, on one host:

```ansi
$ osapi client node process get --target web-01 --pid 1234
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPID   [0m[38;2;10;143;176mNAME   [0m[38;2;10;143;176mUSER   [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mCPU PERCENT  [0m[38;2;10;143;176mMEM PERCENT  [0m[38;2;10;143;176mMEM RSS  [0m[38;2;10;143;176mCOMMAND    [0m[38;2;10;143;176mSTART TIME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      1234  [1msleep[0m  alice  sleep  5.00         1.00         20.0 MB  sleep 300

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node process get --target _all --pid 1234
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPID   [0m[38;2;10;143;176mNAME   [0m[38;2;10;143;176mUSER   [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mCPU PERCENT  [0m[38;2;10;143;176mMEM PERCENT  [0m[38;2;10;143;176mMEM RSS  [0m[38;2;10;143;176mCOMMAND    [0m[38;2;10;143;176mSTART TIME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      1234  [1msleep[0m  alice  sleep  5.00         1.00         20.0 MB  sleep 300
  [1mweb-02[0m    [38;2;103;234;147mok[0m      1234  [1msleep[0m  alice  sleep  5.00         1.00         20.0 MB  sleep 300

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "command": "sleep 300",
      "cpu_percent": 5,
      "hostname": "web-01",
      "mem_percent": 1,
      "mem_rss": 20971520,
      "name": "sleep",
      "pid": 1234,
      "state": "sleep",
      "status": "ok",
      "user": "alice"
    }
  ]
}
```

The command fails in these cases:

| Call                                  | Flags         | Fails with |
| ------------------------------------- | ------------- | ---------- |
| A process that the host does not have | `--pid 99999` | not found  |
| A process ID below 1                  | `--pid -1`    | min        |

## Permission

It needs the `process:read` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--pid`        | The process ID, such as 1234 (required)                |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
