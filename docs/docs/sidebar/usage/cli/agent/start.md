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
| `--agent-queue-group`                 | Queue group for load-balanced (_any) subscriptions (agent.queue_group)                                                       | `job-agents`          |

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-j, --json`       | Enable JSON output                                                                                                      |                         |
| `-f, --osapi-file` | Path to config file                                                                                                     | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
