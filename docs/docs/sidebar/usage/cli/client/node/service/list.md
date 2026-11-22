# List

List the services of the host.

Lists each service that systemd has loaded, with its state, whether it starts at
boot and its description, through gohai. A service that osapi deployed shows by
the name it was deployed with, not its unit's osapi- name, and is managed. A Mac
has no systemd, and reports skipped.

```bash
osapi client node service list [flags]
```

## Examples

```bash
# A service that osapi deployed reads by its name
osapi client node service list --target web-01

# A host with two services
osapi client node service list --target web-01
```

## Result

Each example prints this:

A service that osapi deployed reads by its name:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION                  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh[0m   running  yes      OpenBSD Secure Shell server  no
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m   dead     no       Example app                  yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

A host with two services:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION                  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh[0m        running  yes      OpenBSD Secure Shell server  no
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapi-app[0m  dead     no       Example app                  no

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

## Permission

It needs the `service:read` permission. `osapi token roles get --name <role>`
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
