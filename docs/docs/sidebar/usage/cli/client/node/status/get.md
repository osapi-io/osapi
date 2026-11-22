# Get

Get the status summary of the host.

Runs the get of the hostname, uptime, os, load, memory and disk behaviors on the
host, and answers with each. A read that fails is in errors, as its name and its
error, and the other fields still answer.

```bash
osapi client node status get [flags]
```

## Examples

```bash
# An Ubuntu host
osapi client node status get --target web-01

# A Mac
osapi client node status get --target web-01
```

## Result

Each example prints this:

An Ubuntu host:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME    [0m[38;2;10;143;176mUPTIME                           [0m[38;2;10;143;176mOS INFO       [0m[38;2;10;143;176mLOAD AVERAGE    [0m[38;2;10;143;176mMEMORY                [0m[38;2;10;143;176mDISKS  [0m[38;2;10;143;176mERRORS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mweb-01[0m  1 day, 2 hours, 5 minutes 93900  ubuntu 24.04  0.32 0.28 0.25  8.0 GB 2.0 GB 4.0 GB  /

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A Mac:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME    [0m[38;2;10;143;176mUPTIME                           [0m[38;2;10;143;176mOS INFO      [0m[38;2;10;143;176mLOAD AVERAGE    [0m[38;2;10;143;176mMEMORY                [0m[38;2;10;143;176mDISKS  [0m[38;2;10;143;176mERRORS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mweb-01[0m  1 day, 2 hours, 5 minutes 93900  darwin 15.5  0.32 0.28 0.25  8.0 GB 2.0 GB 4.0 GB  /

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

## Permission

It needs the `status:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
