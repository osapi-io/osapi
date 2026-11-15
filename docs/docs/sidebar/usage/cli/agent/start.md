# Start

Start the agent.

The agent connects to NATS and subscribes to the job streams. It processes each
job when the job arrives. It uses a provider for its platform to run each
operation, for example node status, DNS queries and ping.

```bash
osapi agent start
```

## Examples

```bash
osapi agent start
```

## Global flags

| Flag                         | Description                                                                                                              | Default                 |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ----------------------- |
| `--agent-client-name`        | NATS client name for agent                                                                                               | `osapi-agent`           |
| `--agent-host`               | NATS server hostname for agent                                                                                           | `localhost`             |
| `--agent-hostname`           | Agent hostname (defaults to system hostname)                                                                             |                         |
| `--agent-max-jobs`           | Maximum concurrent jobs per agent                                                                                        | `10`                    |
| `--agent-port`               | NATS server port for agent                                                                                               | `4222`                  |
| `--agent-queue-group`        | NATS queue group for load balancing                                                                                      | `job-agents`            |
| `--consumer-ack-wait`        | Time to wait for acknowledgment before retry                                                                             | `2m`                    |
| `--consumer-back-off`        | Retry backoff intervals                                                                                                  | `[30s,2m,5m,15m,30m]`   |
| `--consumer-max-ack-pending` | Maximum unacknowledged messages                                                                                          | `1000`                  |
| `--consumer-max-deliver`     | Maximum delivery attempts before DLQ                                                                                     | `5`                     |
| `--consumer-replay-policy`   | Replay policy: instant or original                                                                                       | `instant`               |
| `-d, --debug`                | Enable or disable debug mode                                                                                             |                         |
| `-j, --json`                 | Enable JSON output                                                                                                       |                         |
| `-f, --osapi-file`           | Path to config file                                                                                                      | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`               | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key. (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
