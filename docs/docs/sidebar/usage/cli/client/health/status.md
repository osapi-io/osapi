# Status

System status and component health.

The output has: a table of all components, the health of each component, the
NATS connection, statistics for the streams and the KV buckets, the consumers
and the job queue counts. This endpoint needs authentication. The component
table shows all registered components with their health status, active
conditions, uptime, and process resource usage. Sub-components are displayed
nested under their parent with addresses where applicable. Connectivity checks
show the controller's ability to reach NATS and KV dependencies. Use
`osapi client agent list` for agent-specific details like labels and OS info.

```bash
osapi client health status
```

## Examples

```bash
osapi client health status
osapi client health status --json
```

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
