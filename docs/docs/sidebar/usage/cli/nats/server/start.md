# Start

Start the embedded NATS server.

The server makes the streams, consumers and KV buckets that the job system
needs. `osapi start` runs the same server, together with the controller and an
agent, in one process. Use this command to run the bus as its own process. The
`nats.server` section of the config sets the host, the port and the
authentication. Refer to [Configuration](../../../configuration.md).

```bash
osapi nats server start
```

## Examples

```bash
osapi nats server start
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
