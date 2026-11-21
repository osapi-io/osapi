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

It needs the `service:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

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
