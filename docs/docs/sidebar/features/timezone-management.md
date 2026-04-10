---
sidebar_position: 12
---

# Timezone Management

OSAPI manages the system timezone on target hosts via `timedatectl`. The
timezone is read and set atomically, and the update operation is idempotent --
if the timezone already matches the requested value, no change is made and
`changed: false` is returned.

## How It Works

The timezone provider uses `timedatectl` to query and set the system timezone.
The get operation reads the current timezone name and UTC offset. The update
operation compares the current timezone against the requested value and calls
`timedatectl set-timezone` only when a change is needed.

### Get

Runs `timedatectl show -p Timezone --value` to read the IANA timezone name
(e.g., `America/New_York`) and `date +%:z` to read the current UTC offset (e.g.,
`-05:00`).

### Update

Compares the current timezone against the requested value. If they match,
returns `changed: false`. Otherwise, runs `timedatectl set-timezone <timezone>`
and returns `changed: true`.

## Operations

| Operation | Description                                                   |
| --------- | ------------------------------------------------------------- |
| Get       | Get the current timezone name and UTC offset                  |
| Update    | Set the timezone (idempotent, returns changed: false if same) |

## CLI Usage

```bash
# Get timezone from a host
osapi client node timezone get --target web-01

# Update timezone
osapi client node timezone update --target web-01 \
  --timezone America/New_York

# Broadcast update to all hosts
osapi client node timezone update --target _all \
  --timezone UTC
```

All commands support `--json` for raw JSON output.

## Broadcast Support

All timezone operations support broadcast targeting. Use `--target _all` to
query or update every registered agent, or use a label selector like
`--target group:web` to target a subset.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  TIMEZONE          UTC_OFFSET
  web-01    America/New_York  -05:00
  web-02    UTC               +00:00
```

Skipped and failed hosts appear with their error in the output.

## Idempotent Updates

The update operation compares the requested timezone against the current value.
If they match, the response returns `changed: false` and no system call is made.
It is safe to run this operation repeatedly.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, timezone operations return `status: skipped` instead
of failing. See [Platform Detection](../sdk/platform/detection.md) for details
on OS family detection.

## Permissions

| Operation | Permission       |
| --------- | ---------------- |
| Get       | `timezone:read`  |
| Update    | `timezone:write` |

All built-in roles (`admin`, `write`, `read`) include `timezone:read`. The
`admin` and `write` roles also include `timezone:write`.

## Related

- [NTP Management](ntp-management.md) -- NTP server configuration
- [Configuration](../usage/configuration.md) -- full configuration reference
