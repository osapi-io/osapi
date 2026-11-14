# Start

Start the agent:

```bash
$ osapi agent start
```

The agent connects to NATS and subscribes to the job streams. It processes each
job when the job arrives. It uses a provider for its platform to run each
operation, for example node status, DNS queries and ping.

## Configuration

`osapi.yaml` or CLI flags configure the agent:

| Flag                  | Description                         | Default         |
| --------------------- | ----------------------------------- | --------------- |
| `--agent-host`        | NATS server hostname                | `localhost`     |
| `--agent-port`        | NATS server port                    | `4222`          |
| `--agent-client-name` | NATS client name for identification | `osapi-agent`   |
| `--agent-queue-group` | Queue group for load balancing      | `job-agents`    |
| `--agent-hostname`    | Agent hostname for routing          | system hostname |
| `--agent-max-jobs`    | Maximum concurrent jobs             | `10`            |

## Consumer Settings

| Flag                         | Description                         | Default             |
| ---------------------------- | ----------------------------------- | ------------------- |
| `--consumer-max-deliver`     | Max delivery attempts before DLQ    | `5`                 |
| `--consumer-ack-wait`        | Time to wait for acknowledgment     | `2m`                |
| `--consumer-max-ack-pending` | Max unacknowledged messages         | `1000`              |
| `--consumer-replay-policy`   | Replay policy (instant or original) | `instant`           |
| `--consumer-back-off`        | Retry backoff intervals             | `30s,2m,5m,15m,30m` |

## How It Works

1. Connects to NATS and creates JetStream consumers
2. Subscribes to query (`jobs.query.>`) and modify (`jobs.modify.>`) subjects
3. Processes each job: it sends the job to the correct provider
4. Writes status events and results back to the KV store
5. Gracefully shuts down on SIGINT/SIGTERM
