# List

List active agents.

This command reads the agent heartbeat registry directly. It makes no job. Each
agent writes a heartbeat every 10 seconds, with a 30-second TTL. When an agent
stops its heartbeat, it goes from the list.

```bash
osapi client agent list [flags]
```

## Examples

```bash
osapi client agent list
osapi client agent list --pending
```

## Notes

The `--json` output has more system facts that the agent collects: architecture,
kernel version, FQDN, CPU count, network interfaces, service manager and package
manager. The table view does not show these fields.

## Flags

| Flag        | Description                     | Default |
| ----------- | ------------------------------- | ------- |
| `--pending` | List agents awaiting enrollment |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
