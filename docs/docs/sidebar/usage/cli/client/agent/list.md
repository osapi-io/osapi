# List

List active agents.

This command reads the agent heartbeat registry directly. It makes no job. Each
agent writes a heartbeat every 10 seconds, with a 30-second TTL. When an agent
stops its heartbeat, it goes from the list.

```bash
osapi client agent list
```

## Examples

```bash
osapi client agent list
```

## Permission

It needs the `agent:read` permission. `osapi token roles` lists the roles that
grant it.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
