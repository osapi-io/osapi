---
sidebar_position: 1
sidebar_label: Configuration
---

# Configuration

OSAPI is configured through a YAML file and optional environment variable
overrides.

## Config File

By default OSAPI looks for `/etc/osapi/osapi.yaml`. Override the path with the
`-f` / `--osapi-file` flag:

```bash
osapi -f /path/to/osapi.yaml api server start
```

## Environment Variables

Every config key can be overridden with an environment variable using the
`OSAPI_` prefix. Dots and nested keys become underscores, and the name is
uppercased:

| Config Key                         | Environment Variable                     |
| ---------------------------------- | ---------------------------------------- |
| `debug`                            | `OSAPI_DEBUG`                            |
| `api.server.port`                  | `OSAPI_API_SERVER_PORT`                  |
| `api.server.nats.host`             | `OSAPI_API_SERVER_NATS_HOST`             |
| `api.server.nats.port`             | `OSAPI_API_SERVER_NATS_PORT`             |
| `api.server.nats.client_name`      | `OSAPI_API_SERVER_NATS_CLIENT_NAME`      |
| `api.server.nats.namespace`        | `OSAPI_API_SERVER_NATS_NAMESPACE`        |
| `api.server.nats.auth.type`        | `OSAPI_API_SERVER_NATS_AUTH_TYPE`        |
| `api.server.security.signing_key`  | `OSAPI_API_SERVER_SECURITY_SIGNING_KEY`  |
| `api.client.security.bearer_token` | `OSAPI_API_CLIENT_SECURITY_BEARER_TOKEN` |
| `nats.server.host`                 | `OSAPI_NATS_SERVER_HOST`                 |
| `nats.server.port`                 | `OSAPI_NATS_SERVER_PORT`                 |
| `nats.server.namespace`            | `OSAPI_NATS_SERVER_NAMESPACE`            |
| `nats.server.auth.type`            | `OSAPI_NATS_SERVER_AUTH_TYPE`            |
| `nats.stream.name`                 | `OSAPI_NATS_STREAM_NAME`                 |
| `nats.kv.bucket`                   | `OSAPI_NATS_KV_BUCKET`                   |
| `nats.kv.response_bucket`          | `OSAPI_NATS_KV_RESPONSE_BUCKET`          |
| `nats.audit.bucket`                | `OSAPI_NATS_AUDIT_BUCKET`                |
| `nats.audit.ttl`                   | `OSAPI_NATS_AUDIT_TTL`                   |
| `nats.audit.max_bytes`             | `OSAPI_NATS_AUDIT_MAX_BYTES`             |
| `nats.audit.storage`               | `OSAPI_NATS_AUDIT_STORAGE`               |
| `nats.audit.replicas`              | `OSAPI_NATS_AUDIT_REPLICAS`              |
| `telemetry.tracing.enabled`        | `OSAPI_TELEMETRY_TRACING_ENABLED`        |
| `telemetry.tracing.exporter`       | `OSAPI_TELEMETRY_TRACING_EXPORTER`       |
| `telemetry.tracing.otlp_endpoint`  | `OSAPI_TELEMETRY_TRACING_OTLP_ENDPOINT`  |
| `job.worker.nats.host`             | `OSAPI_JOB_WORKER_NATS_HOST`             |
| `job.worker.nats.port`             | `OSAPI_JOB_WORKER_NATS_PORT`             |
| `job.worker.nats.client_name`      | `OSAPI_JOB_WORKER_NATS_CLIENT_NAME`      |
| `job.worker.nats.namespace`        | `OSAPI_JOB_WORKER_NATS_NAMESPACE`        |
| `job.worker.nats.auth.type`        | `OSAPI_JOB_WORKER_NATS_AUTH_TYPE`        |
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

## Authentication

Each NATS connection supports pluggable authentication. Set the `auth.type`
field in the relevant section (`nats.server`, `api.server.nats`, or
`job.worker.nats`):

| Type        | Description                          | Extra Fields           |
| ----------- | ------------------------------------ | ---------------------- |
| `none`      | No authentication (default)          | —                      |
| `user_pass` | Username and password                | `username`, `password` |
| `nkey`      | NKey-based auth (server and clients) | See examples below     |

### Server-Side Auth

The embedded NATS server (`nats.server.auth`) accepts a list of users or nkeys:

```yaml
nats:
  server:
    auth:
      type: user_pass
      users:
        - username: osapi
          password: '<secret>'
```

### Client-Side Auth

API server and worker connections (`api.server.nats.auth`,
`job.worker.nats.auth`) authenticate as a single identity:

```yaml
api:
  server:
    nats:
      auth:
        type: user_pass
        username: osapi
        password: '<secret>'
```

## Permissions

OSAPI uses fine-grained `resource:verb` permissions for access control. Each API
endpoint requires a specific permission. Built-in roles expand to a default set
of permissions:

| Role    | Permissions                                                                                                             |
| ------- | ----------------------------------------------------------------------------------------------------------------------- |
| `admin` | `system:read`, `network:read`, `network:write`, `job:read`, `job:write`, `health:read`, `audit:read`, `command:execute` |
| `write` | `system:read`, `network:read`, `network:write`, `job:read`, `job:write`, `health:read`                                  |
| `read`  | `system:read`, `network:read`, `job:read`, `health:read`                                                                |

### Custom Roles

You can define custom roles in the `api.server.security.roles` section. Custom
roles override the default permission mapping for the same name, or define
entirely new role names:

```yaml
api:
  server:
    security:
      roles:
        ops:
          permissions:
            - system:read
            - health:read
        netadmin:
          permissions:
            - network:read
            - network:write
            - health:read
```

### Direct Permissions

Tokens can carry a `permissions` claim that overrides role-based expansion. When
the claim is present, only the listed permissions are granted regardless of the
token's roles. Generate a token with direct permissions:

```bash
osapi token generate -r admin -u user@example.com \
  -p system:read -p health:read
```

## Namespace

The `namespace` field on NATS connections prefixes all subject names and
infrastructure names. This allows multiple OSAPI deployments to share a single
NATS cluster without collisions.

| Without namespace       | With `namespace: osapi` |
| ----------------------- | ----------------------- |
| `jobs.query._any`       | `osapi.jobs.query._any` |
| `JOBS` (stream)         | `osapi-JOBS`            |
| `job-queue` (KV bucket) | `osapi-job-queue`       |

Set the same namespace value in `nats.server.namespace`,
`api.server.nats.namespace`, and `job.worker.nats.namespace` so all components
agree on naming. An empty string disables prefixing.

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
    nats:
      # NATS server hostname for the API server.
      host: 'localhost'
      # NATS server port for the API server.
      port: 4222
      # Client name sent to NATS for identification.
      client_name: 'osapi-api'
      # Subject namespace prefix. Must match nats.server.namespace.
      namespace: 'osapi'
      auth:
        # Authentication type: "none", "user_pass", or "nkey".
        type: 'none'
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
      # Custom roles with fine-grained permissions.
      # Permissions: system:read, network:read, network:write,
      #              job:read, job:write, health:read, audit:read,
      #              command:execute
      # roles:
      #   ops:
      #     permissions:
      #       - system:read
      #       - health:read

nats:
  server:
    # Hostname the embedded NATS server binds to.
    host: 'localhost'
    # Port the embedded NATS server binds to.
    port: 4222
    # Directory for JetStream file-based storage.
    store_dir: '.nats/jetstream/'
    # Subject namespace prefix for server-created infrastructure.
    namespace: 'osapi'
    auth:
      # Authentication type: "none", "user_pass", or "nkey".
      type: 'none'
      # Users for user_pass auth (server-side only).
      # users:
      #   - username: osapi
      #     password: '<secret>'
      # NKeys for nkey auth (server-side only).
      # nkeys:
      #   - '<public-nkey>'

  # ── JetStream stream ──────────────────────────────────────
  stream:
    # JetStream stream name for job notifications.
    name: 'JOBS'
    # Subject filter for the stream.
    subjects: 'jobs.>'
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

  # ── KV bucket settings ────────────────────────────────────
  kv:
    # KV bucket for immutable job definitions and status events.
    bucket: 'job-queue'
    # KV bucket for worker result storage.
    response_bucket: 'job-responses'
    # TTL for KV entries (Go duration).
    ttl: '1h'
    # Maximum total size of the bucket in bytes.
    max_bytes: 104857600 # 100 MiB
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of KV replicas.
    replicas: 1

  # ── Audit log KV bucket ──────────────────────────────────
  audit:
    # KV bucket for audit log entries.
    bucket: 'audit-log'
    # TTL for audit entries (Go duration). Default 30 days.
    ttl: '720h'
    # Maximum total size of the audit bucket in bytes.
    max_bytes: 52428800 # 50 MiB
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of KV replicas.
    replicas: 1

  # ── Dead Letter Queue ─────────────────────────────────────
  dlq:
    # Maximum age of messages in the DLQ.
    max_age: '7d'
    # Maximum number of messages retained.
    max_msgs: 1000
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of DLQ replicas.
    replicas: 1

telemetry:
  tracing:
    # Enable distributed tracing (default: false).
    enabled: false
    # Exporter type: "stdout" or "otlp".
    # exporter: stdout
    # gRPC endpoint for OTLP exporter (e.g., Jaeger, Tempo).
    # otlp_endpoint: localhost:4317

job:
  worker:
    nats:
      # NATS server hostname for the worker.
      host: 'localhost'
      # NATS server port for the worker.
      port: 4222
      # Client name sent to NATS for identification.
      client_name: 'osapi-job-worker'
      # Subject namespace prefix. Must match nats.server.namespace.
      namespace: 'osapi'
      auth:
        # Authentication type: "none", "user_pass", or "nkey".
        type: 'none'
    consumer:
      # Durable consumer name.
      name: 'jobs-worker'
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
| `nats.host`                   | string   | NATS server hostname                 |
| `nats.port`                   | int      | NATS server port                     |
| `nats.client_name`            | string   | NATS client identification name      |
| `nats.namespace`              | string   | Subject namespace prefix             |
| `nats.auth.type`              | string   | Auth type: `none`, `user_pass`       |
| `nats.auth.username`          | string   | Username for `user_pass` auth        |
| `nats.auth.password`          | string   | Password for `user_pass` auth        |
| `security.signing_key`        | string   | HS256 JWT signing key (**required**) |
| `security.cors.allow_origins` | []string | Allowed CORS origins                 |
| `security.roles`              | map      | Custom roles with permissions lists  |

### `nats.server`

| Key          | Type   | Description                            |
| ------------ | ------ | -------------------------------------- |
| `host`       | string | Hostname the NATS server binds to      |
| `port`       | int    | Port the NATS server binds to          |
| `store_dir`  | string | Directory for JetStream file storage   |
| `namespace`  | string | Namespace prefix for infrastructure    |
| `auth.type`  | string | Auth type: `none`, `user_pass`         |
| `auth.users` | list   | Users for `user_pass` auth (see below) |
| `auth.nkeys` | list   | Public nkeys for `nkey` auth           |

### `nats.stream`

| Key        | Type   | Description                        |
| ---------- | ------ | ---------------------------------- |
| `name`     | string | JetStream stream name              |
| `subjects` | string | Subject filter for the stream      |
| `max_age`  | string | Maximum message age (Go duration)  |
| `max_msgs` | int    | Maximum number of messages         |
| `storage`  | string | `"file"` or `"memory"`             |
| `replicas` | int    | Number of stream replicas          |
| `discard`  | string | Discard policy: `"old"` or `"new"` |

### `nats.kv`

| Key               | Type   | Description                              |
| ----------------- | ------ | ---------------------------------------- |
| `bucket`          | string | KV bucket for job definitions and events |
| `response_bucket` | string | KV bucket for worker results             |
| `ttl`             | string | Entry time-to-live (Go duration)         |
| `max_bytes`       | int    | Maximum bucket size in bytes             |
| `storage`         | string | `"file"` or `"memory"`                   |
| `replicas`        | int    | Number of KV replicas                    |

### `nats.audit`

| Key         | Type   | Description                      |
| ----------- | ------ | -------------------------------- |
| `bucket`    | string | KV bucket for audit log entries  |
| `ttl`       | string | Entry time-to-live (Go duration) |
| `max_bytes` | int    | Maximum bucket size in bytes     |
| `storage`   | string | `"file"` or `"memory"`           |
| `replicas`  | int    | Number of KV replicas            |

### `nats.dlq`

| Key        | Type   | Description                       |
| ---------- | ------ | --------------------------------- |
| `max_age`  | string | Maximum message age (Go duration) |
| `max_msgs` | int    | Maximum number of messages        |
| `storage`  | string | `"file"` or `"memory"`            |
| `replicas` | int    | Number of DLQ replicas            |

### `telemetry.tracing`

| Key             | Type   | Description                                                            |
| --------------- | ------ | ---------------------------------------------------------------------- |
| `enabled`       | bool   | Enable distributed tracing (default: `false`)                          |
| `exporter`      | string | `"stdout"`, `"otlp"`, or unset (log correlation only, no span export)  |
| `otlp_endpoint` | string | gRPC endpoint for OTLP exporter (required when `exporter` is `"otlp"`) |

### `job.worker`

| Key                        | Type              | Description                               |
| -------------------------- | ----------------- | ----------------------------------------- |
| `nats.host`                | string            | NATS server hostname                      |
| `nats.port`                | int               | NATS server port                          |
| `nats.client_name`         | string            | NATS client identification name           |
| `nats.namespace`           | string            | Subject namespace prefix                  |
| `nats.auth.type`           | string            | Auth type: `none`, `user_pass`            |
| `nats.auth.username`       | string            | Username for `user_pass` auth             |
| `nats.auth.password`       | string            | Password for `user_pass` auth             |
| `consumer.name`            | string            | Durable consumer name                     |
| `consumer.max_deliver`     | int               | Max redelivery attempts before DLQ        |
| `consumer.ack_wait`        | string            | ACK timeout (Go duration)                 |
| `consumer.max_ack_pending` | int               | Max outstanding unacknowledged msgs       |
| `consumer.replay_policy`   | string            | `"instant"` or `"original"`               |
| `consumer.back_off`        | []string          | Backoff durations between redeliveries    |
| `queue_group`              | string            | Queue group for load-balanced routing     |
| `hostname`                 | string            | Worker hostname (defaults to OS hostname) |
| `max_jobs`                 | int               | Max concurrent jobs                       |
| `labels`                   | map[string]string | Key-value pairs for label-based routing   |
