# Shutdown

Shut the host down.

Runs shutdown -h -- +N, with N the delay in whole minutes and at least one, so
the agent can answer before the host goes down. A message goes to the logged-in
users. Each run changes the host.

```bash
osapi client node power shutdown [flags]
```

## Examples

```bash
# In one minute
osapi client node power shutdown --target web-01

# In five minutes, with a message
osapi client node power shutdown --target web-01 --delay 300 --message maintenance
```

## Result

In one minute, on one host:

```ansi
$ osapi client node power shutdown --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mACTION    [0m[38;2;10;143;176mDELAY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  shutdown  60

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node power shutdown --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mACTION    [0m[38;2;10;143;176mDELAY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  shutdown  60
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  shutdown  60

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
      "action": "shutdown",
      "changed": true,
      "delay": 60,
      "hostname": "web-01",
      "status": "ok"
    }
  ]
}
```

In five minutes, with a message, on one host:

```ansi
$ osapi client node power shutdown --target web-01 --delay 300 --message maintenance
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mACTION    [0m[38;2;10;143;176mDELAY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  shutdown  300

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node power shutdown --target _all --delay 300 --message maintenance
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mACTION    [0m[38;2;10;143;176mDELAY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  shutdown  300
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  shutdown  300

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
      "action": "shutdown",
      "changed": true,
      "delay": 300,
      "hostname": "web-01",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                         | On the host                          |
| ------------------------------- | ------------------------------------ |
| In one minute                   | runs `shutdown -h -- +1`             |
| In five minutes, with a message | runs `shutdown -h -- +5 maintenance` |

The command fails in these cases:

| Call             | Flags        | Fails with |
| ---------------- | ------------ | ---------- |
| A negative delay | `--delay -1` | Delay      |

## Permission

It needs the `power:execute` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                                 | Default |
| -------------- | --------------------------------------------------------------------------- | ------- |
| `--delay`      | Seconds before the host goes down. It rounds to whole minutes, at least one |         |
| `--message`    | A message to the logged-in users first                                      |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                      | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
