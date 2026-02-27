# Start

Start all OSAPI components in a single process:

```bash
$ osapi start
```

This runs the embedded NATS server, API server, and agent together — the
recommended approach for single-host deployments. All three components start in
order and shut down gracefully on `SIGINT` / `SIGTERM`.

## Startup Order

1. **NATS server** — starts and blocks until ready (5s timeout)
2. **JetStream infrastructure** — creates streams, KV buckets, and DLQ
3. **API server** — begins accepting HTTP requests
4. **Agent** — connects to NATS and starts processing jobs

## Shutdown

Press `Ctrl-C` or send `SIGTERM`. All three components shut down concurrently
within a 10-second deadline, then NATS connections and telemetry exporters are
cleaned up.

## Configuration

The same `osapi.yaml` file configures all three components. See
[Configuration](../../configuration.md) for the full reference.

```bash
$ osapi -f /path/to/osapi.yaml start
```

## When to Use

Use `osapi start` when all three processes run on the same host — the typical
single-host or appliance deployment. For multi-host setups where the NATS
server, API server, and agents run on different machines, start each component
separately:

```bash
# On the NATS host
osapi nats server start

# On the API host
osapi api server start

# On each managed host
osapi agent start
```
