# List

List the entries of the journal.

Runs journalctl --output=json, the newest entries last: the last 100 unless
lines says, of one unit when unit says, since a time when since says, and at a
priority or above when priority says. A priority is its name, such as err or
info. The hand-written log domain read the Debian family only; this reads each
Linux host that has journald.

```bash
osapi client node log list [flags]
```

## Examples

```bash
# The newest entries
osapi client node log list --target web-01

# The last entry of one unit at info or above
osapi client node log list --target web-01 --unit nginx.service --lines 1 --since '1 hour ago' --priority info
```

## Result

The newest entries, on one host:

```ansi
$ osapi client node log list --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTIMESTAMP                    [0m[38;2;10;143;176mUNIT   [0m[38;2;10;143;176mPRIORITY  [0m[38;2;10;143;176mMESSAGE        [0m[38;2;10;143;176mPID[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m2024-03-31T23:50:45.123456Z[0m  nginx  info      Started nginx  1234
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m2024-03-31T23:50:46Z[0m         sshd   notice    Accepted key   5678

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node log list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTIMESTAMP                    [0m[38;2;10;143;176mUNIT   [0m[38;2;10;143;176mPRIORITY  [0m[38;2;10;143;176mMESSAGE        [0m[38;2;10;143;176mPID[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m2024-03-31T23:50:45.123456Z[0m  nginx  info      Started nginx  1234
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m2024-03-31T23:50:46Z[0m         sshd   notice    Accepted key   5678
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m2024-03-31T23:50:45.123456Z[0m  nginx  info      Started nginx  1234
  [38;2;138;138;138mweb-02[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m2024-03-31T23:50:46Z[0m         sshd   notice    Accepted key   5678

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
      "hostname": "web-01",
      "message": "Started nginx",
      "pid": 1234,
      "priority": "info",
      "status": "ok",
      "timestamp": "2024-03-31T23:50:45.123456Z",
      "unit": "nginx"
    },
    {
      "hostname": "web-01",
      "message": "Accepted key",
      "pid": 5678,
      "priority": "notice",
      "status": "ok",
      "timestamp": "2024-03-31T23:50:46Z",
      "unit": "sshd"
    }
  ]
}
```

The last entry of one unit at info or above, on one host:

```ansi
$ osapi client node log list --target web-01 --unit nginx.service --lines 1 --since '1 hour ago' --priority info
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTIMESTAMP                    [0m[38;2;10;143;176mUNIT   [0m[38;2;10;143;176mPRIORITY  [0m[38;2;10;143;176mMESSAGE        [0m[38;2;10;143;176mPID[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m2024-03-31T23:50:45.123456Z[0m  nginx  info      Started nginx  1234

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node log list --target _all --unit nginx.service --lines 1 --since '1 hour ago' --priority info
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTIMESTAMP                    [0m[38;2;10;143;176mUNIT   [0m[38;2;10;143;176mPRIORITY  [0m[38;2;10;143;176mMESSAGE        [0m[38;2;10;143;176mPID[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m2024-03-31T23:50:45.123456Z[0m  nginx  info      Started nginx  1234
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [38;2;138;138;138m2024-03-31T23:50:45.123456Z[0m  nginx  info      Started nginx  1234

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
      "hostname": "web-01",
      "message": "Started nginx",
      "pid": 1234,
      "priority": "info",
      "status": "ok",
      "timestamp": "2024-03-31T23:50:45.123456Z",
      "unit": "nginx"
    }
  ]
}
```

The command fails in these cases:

| Call                                   | Flags                    | Fails with |
| -------------------------------------- | ------------------------ | ---------- |
| Lines below one                        | `--lines -1`             | min        |
| More than 10000 lines                  | `--lines 10001`          | max        |
| A priority that journald does not have | `--priority loud`        | enum       |
| A unit with a space in it              | `--unit 'nginx service'` | unit_name  |

## Permission

It needs the `log:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                                                        | Default |
| -------------- | -------------------------------------------------------------------------------------------------- | ------- |
| `--lines`      | How many of the newest entries; 100 when it is not set                                             |         |
| `--priority`   | Only the entries at this priority or above (emerg, alert, crit, err, warning, notice, info, debug) |         |
| `--since`      | Only the entries since this time, in any form that journalctl --since takes, such as 1 hour ago    |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                                             | `_all`  |
| `--unit`       | Only the entries of this systemd unit, such as nginx.service                                       |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
