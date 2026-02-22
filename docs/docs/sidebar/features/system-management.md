---
sidebar_position: 1
---

# System Management

OSAPI can query a variety of system-level information from managed hosts. All
system operations are read-only today and run through the
[job system](job-system.md), so the API server never needs direct access to the
host.

## What It Manages

| Resource | Description                                        |
| -------- | -------------------------------------------------- |
| Hostname | System hostname                                    |
| Status   | Uptime, OS name and version, kernel, platform info |
| Disk     | Per-mount usage (total, used, free, percent)       |
| Memory   | RAM and swap usage (total, used, free, percent)    |
| Load     | 1-, 5-, and 15-minute load averages                |

## How It Works

System queries are submitted as jobs. The CLI posts a job to the API server, the
API server publishes it to NATS, a worker picks it up and reads the requested
system information, then writes the result back to NATS KV. The CLI polls for
the result and displays it.

You can target a specific host, broadcast to all hosts, or route by label. See
[CLI Reference](../usage/cli/client/system/system.mdx) for usage and examples,
or the [API Reference](/gen/api/system-management-api-system-operations) for the
REST endpoints.

## Configuration

System management uses the general job infrastructure. No domain-specific
configuration is required. See [Configuration](../usage/configuration.md) for
NATS, job worker, and authentication settings.

## Permissions

All system endpoints require the `system:read` permission. The built-in `admin`,
`write`, and `read` roles all include this permission.

## Related

- [CLI Reference](../usage/cli/client/system/system.mdx) -- system commands
- [API Reference](/gen/api/system-management-api-system-operations) -- REST API
  documentation
- [Job System](job-system.md) -- how async job processing works
- [Architecture](../architecture/architecture.md) -- system design overview
