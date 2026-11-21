# Delete

Remove a service that osapi deployed.

Stops and disables the service, removes its unit file and reloads systemd. A
service with no unit file does not change.

```bash
osapi client node service delete [flags]
```

## Examples

```bash
# Remove a service
osapi client node service delete --target web-01 --name app

# A service that osapi did not deploy
osapi client node service delete --target web-01 --name app

# A unit file that osapi did not deploy
osapi client node service delete --target web-01 --name agent
```

## Result

With the default settings, each example does this on the host:

| Example          | On the host                                                                                                                       |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Remove a service | removes `/etc/systemd/system/osapi-app.service`; runs `systemctl disable --now osapi-app.service`; runs `systemctl daemon-reload` |

The command fails in these cases:

| Call                     | Flags          | Fails with               |
| ------------------------ | -------------- | ------------------------ |
| A name with a space      | `--name "a b"` | Name                     |
| A name that is an option | `--name --now` | not a valid service name |

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
