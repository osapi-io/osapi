# Enable

Make a service that osapi deployed start at boot.

Enables the service. A service that is enabled does not change, and nor does a
unit with no \[Install] section, which systemd calls static: it cannot start at
boot. A service that osapi did not deploy is not found.

```bash
osapi client node service enable [flags]
```

## Examples

```bash
# Enable the service
osapi client node service enable --target web-01 --name app

# A unit with no [Install] section
osapi client node service enable --target web-01 --name app
```

## Result

With the default settings, each example does this on the host:

| Example                           | On the host                               |
| --------------------------------- | ----------------------------------------- |
| Enable the service                | runs `systemctl enable osapi-app.service` |
| A unit with no \[Install] section | runs `systemctl enable osapi-app.service` |

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
