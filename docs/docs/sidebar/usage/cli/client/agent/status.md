# Status

Show the health of each agent.

Each agent, with its process CPU and memory and each of its sub-components, from
the health status of the controller. agent list shows the host of each agent.

```bash
osapi client agent status
```

## Examples

```bash
osapi client agent status
osapi client agent status --json
```

## Permission

It needs the `health:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
