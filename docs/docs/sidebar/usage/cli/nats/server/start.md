# Start

Start the embedded NATS server, with JetStream on:

```bash
$ osapi nats server start
```

The server makes the streams, consumers and KV buckets that the job system
needs. `osapi start` runs the same server, together with the controller and an
agent, in one process. Use this command to run the bus as its own process.

The `nats.server` section of the config sets the host, the port and the
authentication. Refer to [Configuration](../../../configuration.md).
