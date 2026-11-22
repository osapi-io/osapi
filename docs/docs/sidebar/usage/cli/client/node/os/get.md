# Get

Get the operating system of the host.

Reads the distribution and its version: the ID and VERSION_ID of /etc/os-release
on Linux, and darwin with the ProductVersion of sw_vers on macOS, such as darwin
15.5.

```bash
osapi client node os get [flags]
```

## Examples

```bash
# An Ubuntu host
osapi client node os get --target web-01

# A Mac
osapi client node os get --target web-01
```

## Result

Each example prints this:

An Ubuntu host:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mOS INFO[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      ubuntu 24.04

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A Mac:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mOS INFO[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      darwin 15.5

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

## Permission

It needs the `os:read` permission. `osapi token roles get --name <role>` shows
what a role grants.

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
