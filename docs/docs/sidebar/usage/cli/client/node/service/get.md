# Get

Get one service.

Reads one service from the same list, by its name with .service or without it. A
service that osapi deployed reads by the name it was deployed with, such as app,
as its writes take it, or by its unit's name, osapi-app. A service that systemd
has not loaded is not found.

```bash
osapi client node service get [flags]
```

## Examples

```bash
# A running service
osapi client node service get --target web-01 --name ssh

# A name with .service
osapi client node service get --target web-01 --name ssh.service

# A service that osapi deployed, by the name it was deployed with
osapi client node service get --target web-01 --name app

# A service that osapi deployed, by its unit's name
osapi client node service get --target web-01 --name osapi-app
```

## Result

Each example prints this:

A running service:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION                  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh[0m   running  yes      OpenBSD Secure Shell server  no

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A name with .service:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION                  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh[0m   running  yes      OpenBSD Secure Shell server  no

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A service that osapi deployed, by the name it was deployed with:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m   dead   no       Example app  yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A service that osapi deployed, by its unit's name:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m   dead   no       Example app  yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

The command fails in these cases:

| Call                                  | Flags           | Fails with |
| ------------------------------------- | --------------- | ---------- |
| A service that systemd has not loaded | `--name nosuch` | not found  |
| A name with a space                   | `--name "a b"`  | Name       |

## Permission

It needs the `service:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                        | Default |
| -------------- | ------------------------------------------------------------------ | ------- |
| `--name`       | The service, such as nginx, with .service or without it (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)             | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
