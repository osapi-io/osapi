# Start

Start all components (NATS, controller, agent).

This runs the embedded NATS server, the API server and the agent together. Use
it for a deployment on one host. For a deployment on more than one host, start
the NATS server, the controller and each agent alone.

The NATS server starts first, and waits until it is ready. Then osapi makes the
streams and buckets, starts the API server and starts the agent. On `SIGINT` or
`SIGTERM`, the three components stop at the same time, in 10 seconds or less.
Then osapi closes the NATS connections and the telemetry exporters.

```bash
osapi start
```

## Examples

```bash
osapi start
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](output.md) tells how the result is printed, and what `--json` gives.
