---
sidebar_position: 23
---

# Log Management

OSAPI provides read-only access to the systemd journal on managed hosts. Log
entries are retrieved via `journalctl` and returned as structured JSON with
timestamp, unit, priority, message, PID, and hostname fields.

## How It Works

The log provider queries the systemd journal on the agent host at request time.
Each request returns a live snapshot — up to the requested number of entries
matching the given filters.

### Query

Returns journal log entries for the target host. The agent runs
`journalctl --output=json` with optional `--lines`, `--since`, and `--priority`
flags and returns the parsed entries.

### QueryUnit

Returns journal log entries scoped to a specific systemd unit. The agent adds
the `-u <unit>` flag to the journalctl invocation alongside the same optional
filters.

### Sources

Returns a sorted list of unique syslog identifiers (log sources) available in
the journal. The agent runs `journalctl --field=SYSLOG_IDENTIFIER`. Use this to
discover what unit names are available before querying.

## Operations

| Operation | Description                                    |
| --------- | ---------------------------------------------- |
| Query     | Query journal entries for the host             |
| QueryUnit | Query journal entries for a specific unit name |
| Sources   | List available log sources (syslog IDs)        |

## CLI Usage

```bash
# Query last 100 log entries on a host (default)
osapi client node log query --target web-01

# Query last 50 error entries in the past hour
osapi client node log query --target web-01 \
  --lines 50 --since 1h --priority err

# Query journal entries for the sshd unit
osapi client node log unit --target web-01 --name sshd.service

# List available log sources
osapi client node log source --target web-01

# Broadcast log query to all hosts
osapi client node log query --target _all --lines 20
```

All commands support `--json` for raw JSON output.

## Broadcast Support

All log operations support broadcast targeting. Use `--target _all` to query
logs on every registered agent, or use a label selector like
`--target group:web` to target a subset.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  TIMESTAMP                  PRIORITY  UNIT          MESSAGE
  2026-01-01T00:00:01+00:00  info      sshd.service  Accepted publickey ...
  2026-01-01T00:00:02+00:00  info      sshd.service  pam_unix(sshd:ses...
```

Skipped and failed hosts appear with their error in the output.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, log operations return `status: skipped` instead of
failing. See [Platform Detection](../sdk/platform/detection.md) for details on
OS family detection.

## Container Behavior

Log operations return `status: skipped` inside containers. `journalctl` requires
a running systemd instance which is not available in standard container
environments.

## Permissions

| Operation                 | Permission |
| ------------------------- | ---------- |
| Query, QueryUnit, Sources | `log:read` |

Log querying requires `log:read`, included in all built-in roles (`admin`,
`write`, `read`).

## Related

- [CLI Reference](../usage/cli/client/node/log/log.md) -- log commands
- [SDK Reference](../sdk/client/operations/log.md) -- Log service
- [Platform Detection](../sdk/platform/detection.md) -- OS family detection
- [Configuration](../usage/configuration.md) -- full configuration reference
