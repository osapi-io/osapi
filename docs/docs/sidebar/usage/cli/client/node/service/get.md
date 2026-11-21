# Get

Get one service.

Reads one service from the same list, by its name with .service or without it. A
service that systemd has not loaded is not found.

```bash
osapi client node service get [flags]
```

## Examples

```bash
# A running service
osapi client node service get --target web-01 --name ssh

# A name with .service
osapi client node service get --target web-01 --name ssh.service
```

## Result

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
