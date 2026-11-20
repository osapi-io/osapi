# Status

System status and component health.

The table has each component that registered: its status, its active conditions,
its uptime, and the CPU and memory of its process.

```bash
osapi client health status
```

## Examples

```bash
osapi client health status
osapi client health status --json
```

## Permission

It needs the `health:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Notes

Each sub-component is under its component, as a tree, with its address and its
URL. The checks tell if the controller can reach NATS and the KV buckets. The
output also has the counts of the job queue, the streams, the KV buckets and the
Object Stores. The command needs a token. `controller status`, `nats status` and
`agent status` show one type of component. Use `osapi client agent list` for the
labels and the OS of each agent. When a check fails, the overall status is
`degraded`, and the check shows the error.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
