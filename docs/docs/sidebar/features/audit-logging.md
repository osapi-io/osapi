---
sidebar_position: 4
---

# Audit logging

OSAPI records a structured audit trail of every API request. Audit entries
capture who made the request, what they did, and how long it took. This provides
accountability and supports compliance requirements.

## How it works

A middleware in the API server records audit entries automatically. Each
authenticated request makes an audit entry. osapi stores it in its own NATS
JetStream stream. No application code needs to log audit events. The middleware
does it.

```mermaid
sequenceDiagram
    participant Client
    participant Middleware as Audit Middleware
    participant Handler
    participant Stream as NATS Stream

    Client->>Middleware: API request
    Middleware->>Handler: forward request
    Handler-->>Middleware: response
    Middleware->>Stream: write audit entry
    Middleware-->>Client: response
```

Each audit entry contains:

| Field             | Description                                      |
| ----------------- | ------------------------------------------------ |
| `id`              | Unique entry identifier (UUID)                   |
| `timestamp`       | When the request was made                        |
| `user`            | Identity from the JWT `sub` claim                |
| `roles`           | Roles from the JWT token                         |
| `method`          | HTTP method (`GET`, `POST`, etc.)                |
| `path`            | Request path (e.g., `/node/hostname`)            |
| `operation_id`    | OpenAPI operation ID (if available)              |
| `source_ip`       | Client IP address                                |
| `response_code`   | HTTP response status code                        |
| `duration_ms`     | Request processing time in milliseconds          |
| `trace_id`        | OpenTelemetry trace ID (when tracing is enabled) |
| `job_id`          | The job the request created, when it created one |
| `request_summary` | What the request asked for, redacted (see below) |

### What the request asked for

The client that called `command/shell` on a host, without the command, does not
tell what happened. Thus, for each method that changes something (`POST`, `PUT`,
`PATCH`, `DELETE`), the entry records `request_summary`. It is the request body
as JSON. osapi replaces the values of sensitive fields with `[redacted]`.

- **Redacted fields.** osapi finds them by name, at all levels, and upper case
  or lower case does not matter. Thus it also redacts a password in a list of
  users: `password`, `password_hash`, `secret`, `token`, `private_key`,
  `key_data`, `stdin`, `content`, `authorization`.
- **The list is a deny list, not an allow list.** osapi records new fields by
  default. A summary with only the fields that somebody remembered does not help
  after an incident. When a change adds a field with a secret, the same change
  adds the field to the list.
- **A summary has a limit** of 2 KB. When osapi cuts it, it adds `…[truncated]`.
  osapi does not read a body of more than 64 KB. The entry says so. It does not
  keep the upload.
- **A body that is not JSON.** osapi records its content type and size. It does
  not store the body.
- **A read records no summary.** A `GET` has nothing that osapi must keep.

`job_id` connects an entry to the status timeline of its job. Thus "who asked
for this" and "what the agents did" are one trail, not two. A request that makes
no job has no `job_id`.

## Viewing audit logs

Query audit entries through the API or CLI. You can list recent entries
(paginated), get a specific entry by ID, or export all entries. Refer to
[CLI Reference](../usage/cli/client/audit/audit.mdx) for usage and examples, or
the [API Reference](/gen/api/audit-log-api-audit) for the REST endpoints.

## Export

Export gets all audit entries in one response, with no pages. Use it to keep
entries for a long time. The audit entries in the NATS stream have a `max_age`
that you can set (30 days by default). An export keeps them before they expire.

Without `--json`, the command shows the entries as a table. To keep them, send
the JSON to a file:

```bash
osapi client audit export --json > audit.json

# Count the entries
jq '.items | length' audit.json

# Find the entries of a user
jq '.items[] | select(.user == "ops@example.com")' audit.json

# Write one entry on each line (JSONL)
jq -c '.items[]' audit.json > audit.jsonl
```

## Retention

osapi stores audit entries in a NATS JetStream stream, with retention settings
that you can set. When entries are older than `max_age`, or when the stream is
at its size limit, NATS removes the oldest entries automatically. If you must
keep entries for a long time, export them before they expire.

## Configuration

```yaml
nats:
  audit:
    stream: 'AUDIT' # JetStream stream name
    subject: 'audit' # Base subject prefix
    max_age: '720h' # 30-day retention (default)
    max_bytes: 52428800 # 50 MiB max stream size
    storage: 'file' # "file" or "memory"
    replicas: 1 # Number of stream replicas
```

When `stream` is empty, the controller keeps no audit log. Each audit endpoint
then answers 503, with a message that names `nats.audit.stream`.

When distributed tracing is enabled, each audit entry has a `trace_id` field.
Use it to find the OpenTelemetry trace of the entry.

Refer to [Configuration](../usage/configuration.md#every-setting) for the full
reference.

## Permissions

All audit endpoints require the `audit:read` permission. Only the `admin` role
includes this permission by default.

A token with a target scope sees only the entries of requests to a target in its
scope. Refer to [Target scope](authentication.md#target-scope).

## Related

- [CLI Reference](../usage/cli/client/audit/audit.mdx): audit commands (list,
  get, export)
- [API Reference](/gen/api/audit-log-api-audit): REST API documentation
- [Configuration](../usage/configuration.md): full configuration reference
