# Get

Get the load averages of the host.

Reads the averages of the run queue over 1, 5 and 15 minutes.

```bash
osapi client node load get [flags]
```

## Examples

```bash
# A host under a light load
osapi client node load get --target web-01
```

## Permission

It needs the `load:read` permission. `osapi token roles get --name <role>` shows
what a role grants.

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
