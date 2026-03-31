---
sidebar_position: 11
---

# NTP Management

OSAPI manages NTP (Network Time Protocol) configuration on target hosts via
chrony. Configuration is written as a drop-in file under `/etc/chrony/conf.d/`
and chrony is reloaded to apply changes immediately. The file-state KV bucket
tracks the SHA-256 of the deployed file so updates are idempotent.

## How It Works

The NTP provider writes a single drop-in file:

```
/etc/chrony/conf.d/osapi-ntp.conf
```

The file contains `server` directives for each configured NTP server address:

```
server 0.pool.ntp.org iburst
server 1.pool.ntp.org iburst
```

After writing the file, chrony is signalled to reload its configuration so
changes take effect without a full daemon restart. The file-state KV bucket
tracks the SHA-256 of the deployed file — if the server list is unchanged, the
file is not rewritten and `changed: false` is returned.

The get operation queries `chronyc tracking` and `chronyc sources` to report
live sync status, stratum, estimated offset, and the current reference source.
It also reads the deployed drop-in file to report the configured server list.

### Delete Behavior

Deleting the NTP configuration removes the drop-in file from
`/etc/chrony/conf.d/osapi-ntp.conf` and signals chrony to reload. The system
will fall back to whatever other chrony configuration remains — typically the
distribution default.

## Operations

| Operation | Description                                           |
| --------- | ----------------------------------------------------- |
| Get       | Get NTP sync status, stratum, offset, and server list |
| Create    | Deploy the drop-in file (fails if already managed)    |
| Update    | Replace the drop-in file (fails if not managed)       |
| Delete    | Remove the drop-in file and reload chrony             |

## CLI Usage

```bash
# Get NTP status from a host
osapi client node ntp get --target web-01

# Create NTP configuration
osapi client node ntp create --target web-01 \
  --servers 0.pool.ntp.org --servers 1.pool.ntp.org

# Update NTP servers
osapi client node ntp update --target web-01 \
  --servers ntp.example.com

# Broadcast create to all hosts
osapi client node ntp create --target _all \
  --servers 0.pool.ntp.org --servers 1.pool.ntp.org

# Delete NTP configuration
osapi client node ntp delete --target web-01
```

All commands support `--json` for raw JSON output.

## Broadcast Support

All NTP operations support broadcast targeting. Use `--target _all` to apply
configuration across every registered agent, or use a label selector like
`--target group:web` to target a subset.

Responses always include per-host results. The get response includes sync status
fields per host:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  SYNCHRONIZED  STRATUM  OFFSET    SOURCE          SERVERS
  web-01    yes           2        +0.000123  192.0.2.1       0.pool.ntp.org, 1.pool.ntp.org
  web-02    yes           2        +0.000045  192.0.2.1       0.pool.ntp.org, 1.pool.ntp.org
```

Skipped and failed hosts appear with their error in the output.

## Idempotent Updates

Both create and update compare the new server list against the SHA-256 stored in
the file-state KV. If the configuration is unchanged, the response returns
`changed: false` and no filesystem write or chrony reload occurs. It is safe to
run these operations repeatedly.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, NTP operations return `status: skipped` instead of
failing. See [Platform Detection](../sdk/platform/detection.md) for details on
OS family detection.

## Permissions

| Operation              | Permission  |
| ---------------------- | ----------- |
| Get                    | `ntp:read`  |
| Create, Update, Delete | `ntp:write` |

All built-in roles (`admin`, `write`, `read`) include `ntp:read`. The `admin`
and `write` roles also include `ntp:write`.

## Related

- [CLI Reference](../usage/cli/client/node/ntp/ntp.md) — NTP commands
- [File Management](file-management.md) — file-state KV tracking
- [Platform Detection](../sdk/platform/detection.md) — OS family detection
- [Configuration](../usage/configuration.md) — full configuration reference
