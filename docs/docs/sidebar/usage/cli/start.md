# Start

Start all components (NATS, controller, agent).

This runs the embedded NATS server, the API server and the agent together. Use
it for a deployment on one host. The three components start in sequence. They
stop cleanly on `SIGINT` / `SIGTERM`. For a deployment on more than one host,
start the NATS server, the controller and each agent alone.

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
