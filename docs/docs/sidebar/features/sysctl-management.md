---
sidebar_position: 10
---

# Sysctl Management

OSAPI manages Linux kernel parameters on target hosts via `sysctl`. Each
parameter is persisted to a drop-in file under `/etc/sysctl.d/` and applied
immediately with `sysctl -p`. This makes changes durable across reboots while
also taking effect instantly.

## How It Works

The sysctl provider writes one file per parameter:

```
/etc/sysctl.d/osapi-{key}.conf
```

For example, setting `net.ipv4.ip_forward` writes:

```
/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf
```

File content is a single `key = value` line:

```
net.ipv4.ip_forward = 1
```

After writing the file, the provider runs
`sysctl -p /etc/sysctl.d/osapi-{key}.conf` to apply the value to the running
kernel immediately. The file-state KV bucket tracks the SHA-256 of each deployed
file so that updates are idempotent — if the value is unchanged, the file is not
rewritten and `changed: false` is returned.

### Delete Behavior

Deleting a sysctl entry removes the drop-in file from
`/etc/sysctl.d/osapi-{key}.conf`. The parameter value is **not** reset in the
running kernel — the active value remains until the next reboot or until you set
it explicitly. To revert a value, set it to its default before deleting the
file.

### List and Get

The list operation queries the file-state KV bucket for all entries deployed by
OSAPI (keys with the `sysctl:` prefix). Manually created files under
`/etc/sysctl.d/` are not visible. The get operation fetches a specific entry by
key from the same KV bucket.

## Operations

| Operation | Description                                            |
| --------- | ------------------------------------------------------ |
| List      | List all OSAPI-managed sysctl parameters               |
| Get       | Get a specific parameter by key                        |
| Create    | Create a new drop-in file and apply (fails if exists)  |
| Update    | Update an existing drop-in file (fails if not managed) |
| Delete    | Remove the drop-in file                                |

## CLI Usage

```bash
# List all managed parameters on a host
osapi client node sysctl list --target web-01

# Get a specific parameter
osapi client node sysctl get --target web-01 \
  --key net.ipv4.ip_forward

# Create a parameter (fails if already managed)
osapi client node sysctl create --target web-01 \
  --key net.ipv4.ip_forward --value 1

# Update an existing parameter
osapi client node sysctl update --target web-01 \
  --key net.ipv4.ip_forward --value 0

# Broadcast create to all hosts
osapi client node sysctl create --target _all \
  --key vm.swappiness --value 10

# Delete a parameter (removes the drop-in file)
osapi client node sysctl delete --target web-01 \
  --key net.ipv4.ip_forward
```

All commands support `--json` for raw JSON output.

## Broadcast Support

All sysctl operations support broadcast targeting. Use `--target _all` to apply
a parameter across every registered agent, or use a label selector like
`--target group:web` to target a subset.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  KEY                   VALUE
  web-01    net.ipv4.ip_forward   1
  web-02    net.ipv4.ip_forward   1
```

Skipped and failed hosts appear with their error in the output.

## Idempotent Updates

Both create and update operations compare the new value against the SHA-256
stored in the file-state KV. If the parameter value is unchanged, the response
returns `changed: false`. It is safe to run these operations repeatedly without
generating unnecessary filesystem writes or `sysctl -p` invocations.

## Key Naming

Sysctl keys follow the standard dot-notation format used by the Linux kernel:
`net.ipv4.ip_forward`, `vm.swappiness`, `kernel.hostname`, etc. OSAPI accepts
any key that sysctl recognizes; validation happens at the agent when the
provider attempts to apply the value.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, sysctl operations return `status: skipped` instead of
failing. See [Platform Detection](../sdk/platform/detection.md) for details on
OS family detection.

## Permissions

| Operation              | Permission     |
| ---------------------- | -------------- |
| List, Get              | `sysctl:read`  |
| Create, Update, Delete | `sysctl:write` |

All built-in roles (`admin`, `write`, `read`) include `sysctl:read`. The `admin`
and `write` roles also include `sysctl:write`.

## Related

- [CLI Reference](../usage/cli/client/node/sysctl/sysctl.md) — sysctl commands
- [File Management](file-management.md) — file-state KV tracking
- [Platform Detection](../sdk/platform/detection.md) — OS family detection
- [Configuration](../usage/configuration.md) — full configuration reference
