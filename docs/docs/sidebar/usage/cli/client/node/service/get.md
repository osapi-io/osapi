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

A running service, on one host:

```ansi
$ osapi client node service get --target web-01 --name ssh
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION                  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh[0m   running  yes      OpenBSD Secure Shell server  no

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node service get --target _all --name ssh
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION                  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh[0m   running  yes      OpenBSD Secure Shell server  no
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mssh[0m   running  yes      OpenBSD Secure Shell server  no

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
      "description": "OpenBSD Secure Shell server",
      "enabled": true,
      "hostname": "web-01",
      "managed": false,
      "name": "ssh",
      "state": "running",
      "status": "ok"
    }
  ]
}
```

A name with .service, on one host:

```ansi
$ osapi client node service get --target web-01 --name ssh.service
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION                  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh[0m   running  yes      OpenBSD Secure Shell server  no

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node service get --target _all --name ssh.service
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE    [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION                  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mssh[0m   running  yes      OpenBSD Secure Shell server  no
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mssh[0m   running  yes      OpenBSD Secure Shell server  no

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
      "description": "OpenBSD Secure Shell server",
      "enabled": true,
      "hostname": "web-01",
      "managed": false,
      "name": "ssh",
      "state": "running",
      "status": "ok"
    }
  ]
}
```

A service that osapi deployed, by the name it was deployed with, on one host:

```ansi
$ osapi client node service get --target web-01 --name app
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m   dead   no       Example app  yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node service get --target _all --name app
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m   dead   no       Example app  yes
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mapp[0m   dead   no       Example app  yes

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
      "description": "Example app",
      "enabled": false,
      "hostname": "web-01",
      "managed": true,
      "name": "app",
      "state": "dead",
      "status": "ok"
    }
  ]
}
```

A service that osapi deployed, by its unit's name, on one host:

```ansi
$ osapi client node service get --target web-01 --name osapi-app
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m   dead   no       Example app  yes

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node service get --target _all --name osapi-app
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME  [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mENABLED  [0m[38;2;10;143;176mDESCRIPTION  [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m   dead   no       Example app  yes
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mapp[0m   dead   no       Example app  yes

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
      "description": "Example app",
      "enabled": false,
      "hostname": "web-01",
      "managed": true,
      "name": "app",
      "state": "dead",
      "status": "ok"
    }
  ]
}
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
