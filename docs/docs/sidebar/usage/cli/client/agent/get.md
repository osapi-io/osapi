# Get

Get agent details.

This command reads the agent heartbeat registry directly. It makes no job. The
data comes from the last heartbeat of the agent. `agent get` shows a small set
of data from the heartbeat registry, at once and with no job. `node status` runs
a full system inspection through the job system, with disk usage and more
metrics.

```bash
osapi client agent get [flags]
```

## Examples

```bash
osapi client agent get --hostname web-01
```

## Flags

| Flag         | Description                                  | Default |
| ------------ | -------------------------------------------- | ------- |
| `--hostname` | Hostname of the agent to retrieve (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
