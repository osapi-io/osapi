# Get

Get the status summary of the host.

Runs the get of the hostname, uptime, os, load, memory and disk behaviors on the
host, and answers with each. A read that fails is in errors, as its name and its
error, and the other fields still answer.

```bash
osapi client node status get [flags]
```

## Examples

```bash
# An Ubuntu host
osapi client node status get --target web-01

# A Mac
osapi client node status get --target web-01
```

## Permission

It needs the `status:read` permission. `osapi token roles get --name <role>`
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
