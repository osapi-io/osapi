# Start

Start the embedded NATS server.

The server makes the streams, consumers and KV buckets that the job system
needs. `osapi start` runs the same server, together with the controller and an
agent, in one process. Use this command to run the bus as its own process. The
`nats.server` section of the config sets the host, the port and the
authentication. Refer to \[Configuration](../../../configuration.md).

```bash
osapi nats server start
```

## Examples

```bash
osapi nats server start
```

## Result

The command prints:

```ansi
$ osapi nats server start
[2m10:00AM[0m [92mINF[0m nats server started successfully [2mcomponent=[0mnats
[2m10:00AM[0m [92mINF[0m embedded NATS server running [2mcomponent=[0mnats [2mhost=[0mlocalhost [2mport=[0m4222 [2mstore_dir=[0m.nats/jetstream/ [2mnamespace=[0mosapi [2mauth.type=[0mnone
[2m10:00AM[0m [92mINF[0m metrics server started [2mcomponent=[0mnats [2msubsystem=[0mmetrics [2maddr=[0m0.0.0.0:9092
[2m10:00AM[0m [92mINF[0m registered in component registry [2mcomponent=[0mnats [2msubsystem=[0mnats.heartbeat [2mhostname=[0mweb-03 [2mtype=[0mnats [2mkey=[0mnats.web-03
```

With `--json`, the command prints:

```json
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"nats server started successfully","component":"nats"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"embedded NATS server running","component":"nats","host":"localhost","port":4222,"store_dir":".nats/jetstream/","namespace":"osapi","auth.type":"none"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"metrics server started","component":"nats","subsystem":"metrics","addr":"0.0.0.0:9092"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"registered in component registry","component":"nats","subsystem":"nats.heartbeat","hostname":"web-03","type":"nats","key":"nats.web-03"}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
