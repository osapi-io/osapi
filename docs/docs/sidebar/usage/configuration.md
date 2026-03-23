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
osapi -f /path/to/osapi.yaml controller start
```

## Environment Variables

Every config key can be overridden with an environment variable using the
`OSAPI_` prefix. Dots and nested keys become underscores, and the name is
uppercased:

| Config Key                                       | Environment Variable                                   |
| ------------------------------------------------ | ------------------------------------------------------ |
| `debug`                                          | `OSAPI_DEBUG`                                          |
| `controller.api.port`                            | `OSAPI_CONTROLLER_API_PORT`                            |
| `controller.api.nats.host`                       | `OSAPI_CONTROLLER_API_NATS_HOST`                       |
| `controller.api.nats.port`                       | `OSAPI_CONTROLLER_API_NATS_PORT`                       |
| `controller.api.nats.client_name`                | `OSAPI_CONTROLLER_API_NATS_CLIENT_NAME`                |
| `controller.api.nats.namespace`                  | `OSAPI_CONTROLLER_API_NATS_NAMESPACE`                  |
| `controller.api.nats.auth.type`                  | `OSAPI_CONTROLLER_API_NATS_AUTH_TYPE`                  |
| `controller.api.security.signing_key`            | `OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEY`            |
| `controller.client.security.bearer_token`        | `OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN`        |
| `controller.metrics.enabled`                     | `OSAPI_CONTROLLER_METRICS_ENABLED`                     |
| `controller.metrics.port`                        | `OSAPI_CONTROLLER_METRICS_PORT`                        |
| `nats.server.host`                               | `OSAPI_NATS_SERVER_HOST`                               |
| `nats.server.port`                               | `OSAPI_NATS_SERVER_PORT`                               |
| `nats.server.namespace`                          | `OSAPI_NATS_SERVER_NAMESPACE`                          |
| `nats.server.auth.type`                          | `OSAPI_NATS_SERVER_AUTH_TYPE`                          |
| `nats.server.metrics.enabled`                    | `OSAPI_NATS_SERVER_METRICS_ENABLED`                    |
| `nats.server.metrics.port`                       | `OSAPI_NATS_SERVER_METRICS_PORT`                       |
| `nats.stream.name`                               | `OSAPI_NATS_STREAM_NAME`                               |
| `nats.kv.bucket`                                 | `OSAPI_NATS_KV_BUCKET`                                 |
| `nats.kv.response_bucket`                        | `OSAPI_NATS_KV_RESPONSE_BUCKET`                        |
| `nats.audit.bucket`                              | `OSAPI_NATS_AUDIT_BUCKET`                              |
| `nats.audit.ttl`                                 | `OSAPI_NATS_AUDIT_TTL`                                 |
| `nats.audit.max_bytes`                           | `OSAPI_NATS_AUDIT_MAX_BYTES`                           |
| `nats.audit.storage`                             | `OSAPI_NATS_AUDIT_STORAGE`                             |
| `nats.audit.replicas`                            | `OSAPI_NATS_AUDIT_REPLICAS`                            |
| `nats.registry.bucket`                           | `OSAPI_NATS_REGISTRY_BUCKET`                           |
| `nats.registry.ttl`                              | `OSAPI_NATS_REGISTRY_TTL`                              |
| `nats.registry.storage`                          | `OSAPI_NATS_REGISTRY_STORAGE`                          |
| `nats.registry.replicas`                         | `OSAPI_NATS_REGISTRY_REPLICAS`                         |
| `nats.facts.bucket`                              | `OSAPI_NATS_FACTS_BUCKET`                              |
| `nats.facts.ttl`                                 | `OSAPI_NATS_FACTS_TTL`                                 |
| `nats.facts.storage`                             | `OSAPI_NATS_FACTS_STORAGE`                             |
| `nats.facts.replicas`                            | `OSAPI_NATS_FACTS_REPLICAS`                            |
| `nats.state.bucket`                              | `OSAPI_NATS_STATE_BUCKET`                              |
| `nats.state.storage`                             | `OSAPI_NATS_STATE_STORAGE`                             |
| `nats.state.replicas`                            | `OSAPI_NATS_STATE_REPLICAS`                            |
| `nats.objects.bucket`                            | `OSAPI_NATS_OBJECTS_BUCKET`                            |
| `nats.objects.max_bytes`                         | `OSAPI_NATS_OBJECTS_MAX_BYTES`                         |
| `nats.objects.storage`                           | `OSAPI_NATS_OBJECTS_STORAGE`                           |
| `nats.objects.replicas`                          | `OSAPI_NATS_OBJECTS_REPLICAS`                          |
| `nats.objects.max_chunk_size`                    | `OSAPI_NATS_OBJECTS_MAX_CHUNK_SIZE`                    |
| `nats.file_state.bucket`                         | `OSAPI_NATS_FILE_STATE_BUCKET`                         |
| `nats.file_state.storage`                        | `OSAPI_NATS_FILE_STATE_STORAGE`                        |
| `nats.file_state.replicas`                       | `OSAPI_NATS_FILE_STATE_REPLICAS`                       |
| `telemetry.tracing.enabled`                      | `OSAPI_TELEMETRY_TRACING_ENABLED`                      |
| `telemetry.tracing.exporter`                     | `OSAPI_TELEMETRY_TRACING_EXPORTER`                     |
| `telemetry.tracing.otlp_endpoint`                | `OSAPI_TELEMETRY_TRACING_OTLP_ENDPOINT`                |
| `controller.notifications.enabled`               | `OSAPI_CONTROLLER_NOTIFICATIONS_ENABLED`               |
| `controller.notifications.notifier`              | `OSAPI_CONTROLLER_NOTIFICATIONS_NOTIFIER`              |
| `controller.notifications.renotify_interval`     | `OSAPI_CONTROLLER_NOTIFICATIONS_RENOTIFY_INTERVAL`     |
| `agent.nats.host`                                | `OSAPI_AGENT_NATS_HOST`                                |
| `agent.nats.port`                                | `OSAPI_AGENT_NATS_PORT`                                |
| `agent.nats.client_name`                         | `OSAPI_AGENT_NATS_CLIENT_NAME`                         |
| `agent.nats.namespace`                           | `OSAPI_AGENT_NATS_NAMESPACE`                           |
| `agent.nats.auth.type`                           | `OSAPI_AGENT_NATS_AUTH_TYPE`                           |
| `agent.hostname`                                 | `OSAPI_AGENT_HOSTNAME`                                 |
| `agent.facts.interval`                           | `OSAPI_AGENT_FACTS_INTERVAL`                           |
| `agent.conditions.memory_pressure_threshold`     | `OSAPI_AGENT_CONDITIONS_MEMORY_PRESSURE_THRESHOLD`     |
| `agent.conditions.high_load_multiplier`          | `OSAPI_AGENT_CONDITIONS_HIGH_LOAD_MULTIPLIER`          |
| `agent.conditions.disk_pressure_threshold`       | `OSAPI_AGENT_CONDITIONS_DISK_PRESSURE_THRESHOLD`       |
| `agent.process_conditions.memory_pressure_bytes` | `OSAPI_AGENT_PROCESS_CONDITIONS_MEMORY_PRESSURE_BYTES` |
| `agent.process_conditions.high_cpu_percent`      | `OSAPI_AGENT_PROCESS_CONDITIONS_HIGH_CPU_PERCENT`      |
| `agent.metrics.enabled`                          | `OSAPI_AGENT_METRICS_ENABLED`                          |
| `agent.metrics.port`                             | `OSAPI_AGENT_METRICS_PORT`                             |

Environment variables take precedence over file values.

## Required Fields

Two fields carry a `required` validation tag and must be set before the server
or client will start:

| Key                                       | Purpose                            |
| ----------------------------------------- | ---------------------------------- |
| `controller.api.security.signing_key`     | HS256 key for signing JWTs         |
| `controller.client.security.bearer_token` | JWT sent with client requests      |
| `controller.metrics.enabled`              | `OSAPI_CONTROLLER_METRICS_ENABLED` |
| `controller.metrics.port`                 | `OSAPI_CONTROLLER_METRICS_PORT`    |

Generate a signing key with `openssl rand -hex 32`. Generate a bearer token with
`osapi token generate`.

## Authentication

Each NATS connection supports pluggable authentication. Set the `auth.type`
field in the relevant section (`nats.server`, `controller.api.nats`, or
`agent.nats`):

| Type        | Description                          | Extra Fields           |
| ----------- | ------------------------------------ | ---------------------- |
| `none`      | No authentication (default)          | —                      |
| `user_pass` | Username and password                | `username`, `password` |
| `nkey`      | NKey-based auth (server and clients) | See examples below     |

### Server-Side Auth

The embedded NATS server (`nats.server.auth`) accepts a list of users or nkeys:

```yaml
nats:
  api:
    auth:
      type: user_pass
      users:
        - username: osapi
          password: '<secret>'
```

### Client-Side Auth

API server and agent connections (`controller.api.nats.auth`, `agent.nats.auth`)
authenticate as a single identity:

```yaml
controller:
  api:
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

| Role    | Permissions                                                                                                                                                                                                                                               |
| ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `admin` | `agent:read`, `agent:write`, `node:read`, `network:read`, `network:write`, `job:read`, `job:write`, `health:read`, `audit:read`, `command:execute`, `file:read`, `file:write`, `docker:read`, `docker:write`, `docker:execute`, `cron:read`, `cron:write` |
| `write` | `agent:read`, `node:read`, `network:read`, `network:write`, `job:read`, `job:write`, `health:read`, `file:read`, `file:write`, `docker:read`, `docker:write`, `cron:read`, `cron:write`                                                                   |
| `read`  | `agent:read`, `node:read`, `network:read`, `job:read`, `health:read`, `file:read`, `docker:read`, `cron:read`                                                                                                                                             |

### Custom Roles

You can define custom roles in the `controller.api.security.roles` section.
Custom roles override the default permission mapping for the same name, or
define entirely new role names:

```yaml
controller:
  api:
    security:
      roles:
        ops:
          permissions:
            - node:read
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
  -p node:read -p health:read
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
`controller.api.nats.namespace`, and `agent.nats.namespace` so all components
agree on naming. An empty string disables prefixing.

## Full Reference

Below is a complete `osapi.yaml` with every supported field and inline comments.
Values shown are representative defaults from the repository's config file.

```yaml
# Enable verbose logging.
debug: true

controller:
  client:
    # Base URL the CLI client connects to.
    url: 'http://0.0.0.0:8080'
    security:
      # JWT bearer token for client requests (REQUIRED).
      # Generate with: osapi token generate
      bearer_token: '<jwt>'

  api:
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
      # Permissions: agent:read, agent:write, node:read, network:read,
      #              network:write, job:read, job:write, health:read,
      #              audit:read, command:execute, file:read, file:write,
      #              docker:read, docker:write, docker:execute,
      #              cron:read, cron:write
      # roles:
      #   ops:
      #     permissions:
      #       - node:read
      #       - health:read

  # Per-component metrics server.
  metrics:
    # Enable the metrics endpoint (default: true).
    enabled: true
    # Port the metrics server listens on.
    port: 9090

  # Condition notification system.
  notifications:
    # Enable the condition watcher and notifier (default: true).
    enabled: true
    # Notifier backend: "log" (default).
    notifier: 'log'
    # How often to re-fire active conditions (Go duration).
    # Zero disables re-notification.
    renotify_interval: '5m'

nats:
  api:
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
    # Per-component metrics server.
    metrics:
      # Enable the metrics endpoint (default: true).
      enabled: true
      # Port the metrics server listens on.
      port: 9092

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
    # KV bucket for agent result storage.
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

  # ── Agent registry KV bucket ──────────────────────────────
  registry:
    # KV bucket for agent heartbeat registration.
    bucket: 'agent-registry'
    # TTL for registry entries (Go duration). Agents refresh
    # every 10s; the TTL acts as a liveness timeout.
    ttl: '30s'
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of KV replicas.
    replicas: 1

  # ── Facts KV bucket ──────────────────────────────────────
  facts:
    # KV bucket for agent facts entries.
    bucket: 'agent-facts'
    # TTL for facts entries (Go duration).
    ttl: '5m'
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of KV replicas.
    replicas: 1

  # ── State KV bucket ──────────────────────────────────────
  state:
    # KV bucket for persistent agent state (drain flags, timeline events).
    # No TTL — operator actions persist indefinitely.
    bucket: 'agent-state'
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of KV replicas.
    replicas: 1

  # ── Object Store (file uploads) ─────────────────────────
  objects:
    # Object Store bucket for uploaded file content.
    bucket: 'file-objects'
    # Maximum total size of the bucket in bytes.
    max_bytes: 104857600 # 100 MiB
    # Storage backend: "file" or "memory".
    storage: 'file'
    # Number of Object Store replicas.
    replicas: 1
    # Maximum chunk size for uploads in bytes.
    max_chunk_size: 262144 # 256 KiB

  # ── File state KV bucket ────────────────────────────────
  file_state:
    # KV bucket for file deploy state tracking.
    # No TTL — state persists until explicitly removed.
    bucket: 'file-state'
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

agent:
  nats:
    # NATS server hostname for the agent.
    host: 'localhost'
    # NATS server port for the agent.
    port: 4222
    # Client name sent to NATS for identification.
    client_name: 'osapi-agent'
    # Subject namespace prefix. Must match nats.server.namespace.
    namespace: 'osapi'
    auth:
      # Authentication type: "none", "user_pass", or "nkey".
      type: 'none'
  consumer:
    # Durable consumer name.
    name: 'jobs-agent'
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
  # Facts collection settings.
  facts:
    # How often the agent collects and publishes facts.
    interval: '60s'
  # Node condition thresholds.
  conditions:
    # Memory pressure threshold (percent used).
    memory_pressure_threshold: 90
    # High load multiplier (load1 / cpu_count).
    high_load_multiplier: 2.0
    # Disk pressure threshold (percent used).
    disk_pressure_threshold: 90
  # Process-level condition thresholds.
  process_conditions:
    # Process RSS threshold in bytes (0 = disabled).
    memory_pressure_bytes: 0
    # Process CPU threshold as percentage (0 = disabled).
    high_cpu_percent: 0
  # Queue group for load-balanced (_any) subscriptions.
  queue_group: 'job-agents'
  # Agent hostname for direct routing. Defaults to the
  # system hostname when empty.
  hostname: ''
  # Maximum number of concurrent jobs to process.
  max_jobs: 10
  # Key-value labels for label-based routing.
  # Values can be hierarchical with dot separators.
  # See Job System Architecture for details.
  labels:
    group: 'web.dev.us-east'
  # Per-component metrics server.
  metrics:
    # Enable the metrics endpoint (default: true).
    enabled: true
    # Port the metrics server listens on.
    port: 9091
```

## Section Reference

### `controller.client`

| Key                     | Type   | Description                        |
| ----------------------- | ------ | ---------------------------------- |
| `url`                   | string | Base URL the CLI client targets    |
| `security.bearer_token` | string | JWT for client auth (**required**) |

### `controller.api`

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

### `controller.metrics`

| Key       | Type | Description                                        |
| --------- | ---- | -------------------------------------------------- |
| `enabled` | bool | Enable the metrics server (default: `true`)        |
| `port`    | int  | Port the metrics server listens on (default: 9090) |

When enabled, the port also serves `/health` (liveness) and `/health/ready`
(readiness) probes without authentication.

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

### `nats.server.metrics`

| Key       | Type | Description                                        |
| --------- | ---- | -------------------------------------------------- |
| `enabled` | bool | Enable the metrics server (default: `true`)        |
| `port`    | int  | Port the metrics server listens on (default: 9092) |

When enabled, the port also serves `/health` (liveness) and `/health/ready`
(readiness) probes without authentication.

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
| `response_bucket` | string | KV bucket for agent results              |
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

### `nats.registry`

| Key        | Type   | Description                           |
| ---------- | ------ | ------------------------------------- |
| `bucket`   | string | KV bucket for agent heartbeat entries |
| `ttl`      | string | Entry time-to-live / liveness timeout |
| `storage`  | string | `"file"` or `"memory"`                |
| `replicas` | int    | Number of KV replicas                 |

### `nats.facts`

| Key        | Type   | Description                       |
| ---------- | ------ | --------------------------------- |
| `bucket`   | string | KV bucket for agent facts entries |
| `ttl`      | string | Entry time-to-live (Go duration)  |
| `storage`  | string | `"file"` or `"memory"`            |
| `replicas` | int    | Number of KV replicas             |

### `nats.state`

| Key        | Type   | Description                                   |
| ---------- | ------ | --------------------------------------------- |
| `bucket`   | string | KV bucket for persistent agent state (no TTL) |
| `storage`  | string | `"file"` or `"memory"`                        |
| `replicas` | int    | Number of KV replicas                         |

### `nats.objects`

| Key              | Type   | Description                          |
| ---------------- | ------ | ------------------------------------ |
| `bucket`         | string | Object Store bucket for file uploads |
| `max_bytes`      | int    | Maximum bucket size in bytes         |
| `storage`        | string | `"file"` or `"memory"`               |
| `replicas`       | int    | Number of Object Store replicas      |
| `max_chunk_size` | int    | Maximum chunk size for uploads       |

### `nats.file_state`

| Key        | Type   | Description                              |
| ---------- | ------ | ---------------------------------------- |
| `bucket`   | string | KV bucket for file deploy state (no TTL) |
| `storage`  | string | `"file"` or `"memory"`                   |
| `replicas` | int    | Number of KV replicas                    |

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

### `controller.notifications`

| Key                 | Type   | Description                                                             |
| ------------------- | ------ | ----------------------------------------------------------------------- |
| `enabled`           | bool   | Enable the condition watcher and notifier (default: `false`)            |
| `notifier`          | string | Notification backend: `"log"` writes condition events to the server log |
| `renotify_interval` | string | Re-fire interval for active conditions (Go duration, default: `"0"`)    |

### `agent`

| Key                                        | Type              | Description                                                |
| ------------------------------------------ | ----------------- | ---------------------------------------------------------- |
| `nats.host`                                | string            | NATS server hostname                                       |
| `nats.port`                                | int               | NATS server port                                           |
| `nats.client_name`                         | string            | NATS client identification name                            |
| `nats.namespace`                           | string            | Subject namespace prefix                                   |
| `nats.auth.type`                           | string            | Auth type: `none`, `user_pass`                             |
| `nats.auth.username`                       | string            | Username for `user_pass` auth                              |
| `nats.auth.password`                       | string            | Password for `user_pass` auth                              |
| `consumer.name`                            | string            | Durable consumer name                                      |
| `consumer.max_deliver`                     | int               | Max redelivery attempts before DLQ                         |
| `consumer.ack_wait`                        | string            | ACK timeout (Go duration)                                  |
| `consumer.max_ack_pending`                 | int               | Max outstanding unacknowledged msgs                        |
| `consumer.replay_policy`                   | string            | `"instant"` or `"original"`                                |
| `consumer.back_off`                        | []string          | Backoff durations between redeliveries                     |
| `queue_group`                              | string            | Queue group for load-balanced routing                      |
| `hostname`                                 | string            | Agent hostname (defaults to OS hostname)                   |
| `max_jobs`                                 | int               | Max concurrent jobs                                        |
| `facts.interval`                           | string            | How often the agent collects facts                         |
| `conditions.memory_pressure_threshold`     | int               | Memory pressure threshold percent (default 90)             |
| `conditions.high_load_multiplier`          | float             | Load multiplier over CPU count (default 2.0)               |
| `conditions.disk_pressure_threshold`       | int               | Disk pressure threshold percent (default 90)               |
| `process_conditions.memory_pressure_bytes` | int64             | Process RSS threshold in bytes (0 = disabled)              |
| `process_conditions.high_cpu_percent`      | float             | Process CPU usage threshold as a percentage (0 = disabled) |
| `labels`                                   | map[string]string | Key-value pairs for label-based routing                    |
| `metrics.enabled`                          | bool              | Enable the metrics server (default: true)                  |
| `metrics.port`                             | int               | Port the metrics server listens on (default: 9091)         |

When `metrics.enabled` is true, the port also serves `/health` (liveness) and
`/health/ready` (readiness) probes without authentication.
