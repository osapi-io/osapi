---
sidebar_position: 23
---

# Log management

osapi gives read-only access to the systemd journal on managed hosts. It gets
the log entries with `journalctl`, and returns them as structured JSON with the
timestamp, unit, priority, message, PID and hostname fields.

## How it works

The log provider reads the systemd journal on the agent host when the request
arrives. Each request returns the current entries that match the filters, up to
the requested number.

### Query

Returns journal log entries for the target host. The agent runs
`journalctl --output=json` with optional `--lines`, `--since`, and `--priority`
flags and returns the parsed entries.

### QueryUnit

Returns journal log entries scoped to a specific systemd unit. The agent adds
the `-u <unit>` flag to the journalctl invocation alongside the same optional
filters.

### Sources

Returns a sorted list of the different syslog identifiers (log sources) in the
journal. The agent runs `journalctl --field=SYSLOG_IDENTIFIER`. Use it to find
the unit names before you query.

## Operations

| Operation | Description                                    |
| --------- | ---------------------------------------------- |
| Query     | Query journal entries for the host             |
| QueryUnit | Query journal entries for a specific unit name |
| Sources   | List available log sources (syslog IDs)        |

## CLI usage

```bash
# Query last 100 log entries on a host (default)
osapi client node log list --target web-01

# Query last 50 error entries in the past hour
osapi client node log list --target web-01 \
  --lines 50 --since 1h --priority err

# Query journal entries for the sshd unit
osapi client node log list --target web-01 --unit sshd.service

# List available log sources
osapi client node log source list --target web-01

# List the logs of each host
osapi client node log list --target _all --lines 20
```

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All log operations accept broadcast targets. Use `--target _all` to query the
logs on all registered agents. Use a label selector, for example
`--target group:web`, to target a part of the fleet.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  TIMESTAMP                  PRIORITY  UNIT          MESSAGE
  2026-01-01T00:00:01+00:00  info      sshd.service  Accepted publickey ...
  2026-01-01T00:00:02+00:00  info      sshd.service  pam_unix(sshd:ses...
```

Skipped and failed hosts appear with their error in the output.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, log operations return `status: skipped`.
They do not fail. Refer to [Platform Detection](../sdk/platform/detection.md)
for how osapi finds the OS family.

## Container behavior

Log operations return `status: skipped` inside containers. `journalctl` requires
a running systemd instance which is not available in standard container
environments.

## Permissions

| Operation                 | Permission |
| ------------------------- | ---------- |
| Query, QueryUnit, Sources | `log:read` |

To query logs, you need `log:read`. All built-in roles (`admin`, `write`,
`read`) have it.

## Related

- [CLI Reference](../usage/cli/client/node/log/log.mdx): log commands
- [SDK Reference](../sdk/client/node/log.md): Log service
- [Platform Detection](../sdk/platform/detection.md): OS family detection
- [Configuration](../usage/configuration.md): full configuration reference
