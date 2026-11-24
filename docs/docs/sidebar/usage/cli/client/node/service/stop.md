# Stop

Stop a service that osapi deployed.

Stops the service. A service that does not run does not change. A service that
osapi did not deploy is not found.

```bash
osapi client node service stop [flags]
```

## Examples

```bash
# Stop the service
osapi client node service stop --target web-01 --name app
```

## Result

Stop the service, on one host:

```ansi
$ osapi client node service stop --target web-01 --name app
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node service stop --target _all --name app
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m

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
      "name": "app",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example          | On the host                             |
| ---------------- | --------------------------------------- |
| Stop the service | runs `systemctl stop osapi-app.service` |

The command fails in these cases:

| Call                                  | Flags          | Fails with               |
| ------------------------------------- | -------------- | ------------------------ |
| A service that osapi did not deploy   | `--name ssh`   | not found                |
| A unit file that osapi did not deploy | `--name agent` | not found                |
| A name with a space                   | `--name "a b"` | Name                     |
| A name that is an option              | `--name --now` | not a valid service name |

## Permission

It needs the `service:execute` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                               | Default |
| -------------- | --------------------------------------------------------- | ------- |
| `--name`       | The service, such as app for osapi-app.service (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)    | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
