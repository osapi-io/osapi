# Start

Start the node agent.

The agent connects to NATS and subscribes to the job streams. It processes each
job when the job arrives. It uses a provider for its platform to run each
operation, for example node status, DNS queries and ping.

```bash
osapi agent start [flags]
```

## Examples

```bash
osapi agent start
osapi agent start --agent-max-jobs 20
```

## Result

The command prints:

```ansi
$ osapi agent start
[2m10:00AM[0m [92mINF[0m starting node agent [2mcomponent=[0magent [2msubsystem=[0magent
[2m10:00AM[0m [92mINF[0m agent configuration [2mcomponent=[0magent [2msubsystem=[0magent [2mmachine_id=[0md530664edc1c431eb5ef1e9dfd6617a5 [2mhostname=[0mweb-03 [2mmax_jobs=[0m10 [2mlabels=[0mmap[group:web.dev.us-east]
[2m10:00AM[0m [92mINF[0m registered in agent registry [2mcomponent=[0magent [2msubsystem=[0magent.heartbeat [2mmachine_id=[0md530664edc1c431eb5ef1e9dfd6617a5 [2mhostname=[0mweb-03 [2mkey=[0magents.d530664edc1c431eb5ef1e9dfd6617a5 [2mttl=[0m30s
[2m10:00AM[0m [92mINF[0m node agent started successfully [2mcomponent=[0magent [2msubsystem=[0magent
[2m10:00AM[0m [92mINF[0m metrics server started [2mcomponent=[0magent [2msubsystem=[0mmetrics [2maddr=[0m0.0.0.0:9091
[2m10:00AM[0m [92mINF[0m http(s) server started [2mcomponent=[0magent [2msubsystem=[0mmetrics [2maddress=[0m[::]:9091
```

With `--json`, the command prints:

```json
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"starting node agent","component":"agent","subsystem":"agent"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"agent configuration","component":"agent","subsystem":"agent","machine_id":"d530664edc1c431eb5ef1e9dfd6617a5","hostname":"web-03","max_jobs":10,"labels":{"group":"web.dev.us-east"}}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"registered in agent registry","component":"agent","subsystem":"agent.heartbeat","machine_id":"d530664edc1c431eb5ef1e9dfd6617a5","hostname":"web-03","key":"agents.d530664edc1c431eb5ef1e9dfd6617a5","ttl":"30s"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"node agent started successfully","component":"agent","subsystem":"agent"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"metrics server started","component":"agent","subsystem":"metrics","addr":"0.0.0.0:9091"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"http(s) server started","component":"agent","subsystem":"metrics","address":"[::]:9091"}
```

## Flags

| Flag                                  | Description                                                                                                                  | Default               |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | --------------------- |
| `--agent-consumer-ack-wait`           | Time to wait for an ACK before redelivering (agent.consumer.ack_wait)                                                        | `2m`                  |
| `--agent-consumer-back-off`           | Backoff durations between redelivery attempts (agent.consumer.back_off)                                                      | `[30s,2m,5m,15m,30m]` |
| `--agent-consumer-inactive-threshold` | How long a job consumer can be idle before NATS removes it (Go duration). Empty keeps it (agent.consumer.inactive_threshold) | `24h`                 |
| `--agent-consumer-max-ack-pending`    | Maximum outstanding unacknowledged messages (agent.consumer.max_ack_pending)                                                 | `1000`                |
| `--agent-consumer-max-deliver`        | Maximum redelivery attempts before sending to DLQ (agent.consumer.max_deliver)                                               | `5`                   |
| `--agent-consumer-replay-policy`      | Replay policy: 'instant' or 'original' (agent.consumer.replay_policy)                                                        | `instant`             |
| `--agent-hostname`                    | Agent hostname for direct routing. Defaults to the system hostname when empty (agent.hostname)                               |                       |
| `--agent-max-jobs`                    | Maximum number of concurrent jobs to process (agent.max_jobs)                                                                | `10`                  |
| `--agent-nats-client-name`            | Client name sent to NATS for identification (agent.nats.client_name)                                                         | `osapi-agent`         |
| `--agent-nats-host`                   | NATS server hostname to connect to (agent.nats.host)                                                                         | `localhost`           |
| `--agent-nats-port`                   | NATS server port to connect to (agent.nats.port)                                                                             | `4222`                |

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-j, --json`       | Enable JSON output                                                                                                      |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults                                        | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
