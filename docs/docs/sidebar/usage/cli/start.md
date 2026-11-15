# Start

Start all OSAPI components in a single process:

```bash
$ osapi start
```

This runs the embedded NATS server, the API server and the agent together. Use
it for a deployment on one host. The three components start in sequence. They
stop cleanly on `SIGINT` / `SIGTERM`.

## Startup Order

1. **NATS server**: it starts, and waits until it is ready (5s timeout)
2. **JetStream infrastructure**: it makes the streams, the KV buckets and the
   DLQ
3. **API server**: it starts to accept HTTP requests
4. **Agent**: it connects to NATS and starts to process jobs

## Shutdown

Press `Ctrl-C` or send `SIGTERM`. The three components stop at the same time, in
10 seconds or less. Then osapi closes the NATS connections and the telemetry
exporters.

## Configuration

The same `osapi.yaml` file configures all three components. Refer to
[Configuration](../configuration.md) for the full reference.

```bash
$ osapi -f /path/to/osapi.yaml start
```

## When to Use

Use `osapi start` when the three processes run on the same host, for example on
one host or on an appliance. For multi-host setups where the NATS server, API
server, and agents run on different machines, start each component separately:

```bash
# On the NATS host
osapi nats server start

# On the controller host
osapi controller start

# On each managed host
osapi agent start
```
