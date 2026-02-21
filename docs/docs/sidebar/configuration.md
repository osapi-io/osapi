---
sidebar_position: 6
---

# Configuration

OSAPI is configured through a YAML file and optional environment variable
overrides.

## Config File

By default OSAPI looks for `osapi.yaml` in the current working directory.
Override the path with the `-f` / `--osapi-file` flag:

```bash
osapi --osapi-file /etc/osapi/osapi.yaml api server start
```

## Environment Variables

Every config key can be overridden with an environment variable using the
`OSAPI_` prefix. Dots and nested keys become underscores, and the name is
uppercased:

| Config Key                         | Environment Variable                     |
| ---------------------------------- | ---------------------------------------- |
| `debug`                            | `OSAPI_DEBUG`                            |
| `api.server.port`                  | `OSAPI_API_SERVER_PORT`                  |
| `api.server.security.signing_key`  | `OSAPI_API_SERVER_SECURITY_SIGNING_KEY`  |
| `api.client.security.bearer_token` | `OSAPI_API_CLIENT_SECURITY_BEARER_TOKEN` |
| `nats.server.host`                 | `OSAPI_NATS_SERVER_HOST`                 |
| `telemetry.tracing.enabled`        | `OSAPI_TELEMETRY_TRACING_ENABLED`        |
| `telemetry.tracing.exporter`       | `OSAPI_TELEMETRY_TRACING_EXPORTER`       |
| `telemetry.tracing.otlp_endpoint`  | `OSAPI_TELEMETRY_TRACING_OTLP_ENDPOINT`  |
| `job.worker.hostname`              | `OSAPI_JOB_WORKER_HOSTNAME`              |

Environment variables take precedence over file values.

## Required Fields

Two fields carry a `required` validation tag and must be set before the server
or client will start:

| Key                                | Purpose                       |
| ---------------------------------- | ----------------------------- |
| `api.server.security.signing_key`  | HS256 key for signing JWTs    |
| `api.client.security.bearer_token` | JWT sent with client requests |

Generate a signing key with `openssl rand -hex 32`. Generate a bearer token with
`osapi token generate`.

## Full Reference

Below is a complete `osapi.yaml` with every supported field and inline comments.
Values shown are representative defaults from the repository's config file.

```yaml
# Enable verbose logging.
debug: true

api:
  client:
    # Base URL the CLI client connects to.
    url: 'http://0.0.0.0:8080'
    security:
      # JWT bearer token for client requests (REQUIRED).
      # Generate with: osapi token generate
      bearer_token: '<jwt>'

  server:
    # Port the REST API server listens on.
    port: 8080
    security:
      # HS256 signing key for JWT validation (REQUIRED).
      # Generate with: openssl rand -hex 32
      signing_key: '<secret>'
      cors:
        # Origins allowed to make cross-origin requests.
        # An empty list disables CORS headers entirely.
        allow_origins:
          - 'http://localhost:3001'
          - 'https://osapi-io.github.io'

nats:
  server:
    # Hostname the embedded NATS server binds to.
    host: 'localhost'
    # Port the embedded NATS server binds to.
    port: 4222
    # Directory for JetStream file-based storage.
    store_dir: '.nats/jetstream/'

telemetry:
  tracing:
    # Enable distributed tracing (default: false).
    enabled: false
    # Exporter type: "stdout" or "otlp".
    # exporter: stdout
    # gRPC endpoint for OTLP exporter (e.g., Jaeger, Tempo).
    # otlp_endpoint: localhost:4317

job:
  # ── Shared infrastructure names ──────────────────────────
  # JetStream stream name for job notifications.
  stream_name: 'JOBS'
  # Subject filter for the JOBS stream.
  stream_subjects: 'jobs.>'
  # KV bucket for immutable job definitions and status events.
  kv_bucket: 'job-queue'
  # KV bucket for worker result storage.
  kv_response_bucket: 'job-responses'
  # Durable consumer name.
  consumer_name: 'jobs-worker'

  # ── Stream settings ─────────────────────────────────────
  stream:
    # Maximum age of messages in the stream (Go duration).
    max_age: '24h'
    # Maximum number of messages retained.
    max_msgs: 10000
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of stream replicas (1 for single-node).
    replicas: 1
    # Discard policy when limits are reached: "old" or "new".
    discard: 'old'

  # ── Consumer settings ───────────────────────────────────
  consumer:
    # Maximum redelivery attempts before sending to DLQ.
    max_deliver: 5
    # Time to wait for an ACK before redelivering.
    ack_wait: '2m'
    # Maximum outstanding unacknowledged messages.
    max_ack_pending: 1000
    # Replay policy: "instant" or "original".
    replay_policy: 'instant'
    # Backoff durations between redelivery attempts.
    back_off:
      - '30s'
      - '2m'
      - '5m'
      - '15m'
      - '30m'

  # ── KV bucket settings ─────────────────────────────────
  kv:
    # TTL for KV entries (Go duration).
    ttl: '1h'
    # Maximum total size of the bucket in bytes.
    max_bytes: 104857600 # 100 MiB
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of KV replicas.
    replicas: 1

  # ── Dead Letter Queue settings ──────────────────────────
  dlq:
    # Maximum age of messages in the DLQ.
    max_age: '7d'
    # Maximum number of messages retained.
    max_msgs: 1000
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of DLQ replicas.
    replicas: 1

  # ── Job client (CLI → NATS) ────────────────────────────
  client:
    # NATS server hostname for the job client.
    host: 'localhost'
    # NATS server port for the job client.
    port: 4222
    # Client name sent to NATS for identification.
    client_name: 'osapi-jobs-cli'

  # ── Job worker ─────────────────────────────────────────
  worker:
    # NATS server hostname for the worker.
    host: 'localhost'
    # NATS server port for the worker.
    port: 4222
    # Client name sent to NATS for identification.
    client_name: 'osapi-job-worker'
    # Queue group for load-balanced (_any) subscriptions.
    queue_group: 'job-workers'
    # Worker hostname for direct routing. Defaults to the
    # system hostname when empty.
    hostname: ''
    # Maximum number of concurrent jobs to process.
    max_jobs: 10
    # Key-value labels for label-based routing.
    # Values can be hierarchical with dot separators.
    # See Job System Architecture for details.
    labels:
      group: 'web.dev.us-east'
```

## Section Reference

### `api.client`

| Key                     | Type   | Description                        |
| ----------------------- | ------ | ---------------------------------- |
| `url`                   | string | Base URL the CLI client targets    |
| `security.bearer_token` | string | JWT for client auth (**required**) |

### `api.server`

| Key                           | Type     | Description                          |
| ----------------------------- | -------- | ------------------------------------ |
| `port`                        | int      | Port the API server listens on       |
| `security.signing_key`        | string   | HS256 JWT signing key (**required**) |
| `security.cors.allow_origins` | []string | Allowed CORS origins                 |

### `nats.server`

| Key         | Type   | Description                          |
| ----------- | ------ | ------------------------------------ |
| `host`      | string | Hostname the NATS server binds to    |
| `port`      | int    | Port the NATS server binds to        |
| `store_dir` | string | Directory for JetStream file storage |

### `telemetry.tracing`

| Key             | Type   | Description                                   |
| --------------- | ------ | --------------------------------------------- |
| `enabled`       | bool   | Enable distributed tracing (default: `false`) |
| `exporter`      | string | Exporter type: `"stdout"` or `"otlp"`         |
| `otlp_endpoint` | string | gRPC endpoint for OTLP exporter               |

### `job` (top-level)

| Key                  | Type   | Description                              |
| -------------------- | ------ | ---------------------------------------- |
| `stream_name`        | string | JetStream stream name                    |
| `stream_subjects`    | string | Subject filter for the stream            |
| `kv_bucket`          | string | KV bucket for job definitions and events |
| `kv_response_bucket` | string | KV bucket for worker results             |
| `consumer_name`      | string | Durable consumer name                    |

### `job.stream`

| Key        | Type   | Description                        |
| ---------- | ------ | ---------------------------------- |
| `max_age`  | string | Maximum message age (Go duration)  |
| `max_msgs` | int    | Maximum number of messages         |
| `storage`  | string | `"file"` or `"memory"`             |
| `replicas` | int    | Number of stream replicas          |
| `discard`  | string | Discard policy: `"old"` or `"new"` |

### `job.consumer`

| Key               | Type     | Description                            |
| ----------------- | -------- | -------------------------------------- |
| `max_deliver`     | int      | Max redelivery attempts before DLQ     |
| `ack_wait`        | string   | ACK timeout (Go duration)              |
| `max_ack_pending` | int      | Max outstanding unacknowledged msgs    |
| `replay_policy`   | string   | `"instant"` or `"original"`            |
| `back_off`        | []string | Backoff durations between redeliveries |

### `job.kv`

| Key         | Type   | Description                      |
| ----------- | ------ | -------------------------------- |
| `ttl`       | string | Entry time-to-live (Go duration) |
| `max_bytes` | int    | Maximum bucket size in bytes     |
| `storage`   | string | `"file"` or `"memory"`           |
| `replicas`  | int    | Number of KV replicas            |

### `job.dlq`

| Key        | Type   | Description                       |
| ---------- | ------ | --------------------------------- |
| `max_age`  | string | Maximum message age (Go duration) |
| `max_msgs` | int    | Maximum number of messages        |
| `storage`  | string | `"file"` or `"memory"`            |
| `replicas` | int    | Number of DLQ replicas            |

### `job.client`

| Key           | Type   | Description                     |
| ------------- | ------ | ------------------------------- |
| `host`        | string | NATS server hostname            |
| `port`        | int    | NATS server port                |
| `client_name` | string | NATS client identification name |

### `job.worker`

| Key           | Type              | Description                               |
| ------------- | ----------------- | ----------------------------------------- |
| `host`        | string            | NATS server hostname                      |
| `port`        | int               | NATS server port                          |
| `client_name` | string            | NATS client identification name           |
| `queue_group` | string            | Queue group for load-balanced routing     |
| `hostname`    | string            | Worker hostname (defaults to OS hostname) |
| `max_jobs`    | int               | Max concurrent jobs                       |
| `labels`      | map[string]string | Key-value pairs for label-based routing   |
