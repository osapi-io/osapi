# Start

Start the node agent:

```bash
$ osapi node agent start
```

The agent connects to NATS, subscribes to job streams, and processes jobs as
they become available. It uses platform-specific providers to execute operations
(node status, DNS queries, ping, etc.).

## Configuration

Agent behavior is configured via `osapi.yaml` or CLI flags:

| Flag                   | Description                         | Default            |
| ---------------------- | ----------------------------------- | ------------------ |
| `--worker-host`        | NATS server hostname                | `localhost`        |
| `--worker-port`        | NATS server port                    | `4222`             |
| `--worker-client-name` | NATS client name for identification | `osapi-job-worker` |
| `--worker-queue-group` | Queue group for load balancing      | `job-workers`      |
| `--worker-hostname`    | Agent hostname for routing          | system hostname    |
| `--worker-max-jobs`    | Maximum concurrent jobs             | `10`               |

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
3. Processes jobs by dispatching to the appropriate provider
4. Writes status events and results back to the KV store
5. Gracefully shuts down on SIGINT/SIGTERM
