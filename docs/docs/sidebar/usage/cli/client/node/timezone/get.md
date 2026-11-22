# Get

Get the timezone of the host.

Reads the IANA name of the timezone and its offset from UTC now: the zone that
/etc/localtime links to, or /etc/timezone on Linux.

```bash
osapi client node timezone get [flags]
```

## Examples

```bash
# A Linux host on UTC
osapi client node timezone get --target web-01

# A Mac on UTC
osapi client node timezone get --target web-01
```

## Result

Each example prints this:

A Linux host on UTC:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTIMEZONE  [0m[38;2;10;143;176mUTC OFFSET[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      UTC       +00:00

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A Mac on UTC:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mTIMEZONE  [0m[38;2;10;143;176mUTC OFFSET[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      UTC       +00:00

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

## Permission

It needs the `timezone:read` permission. `osapi token roles get --name <role>`
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
