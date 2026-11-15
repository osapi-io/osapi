---
sidebar_position: 12
---

# Timezone management

osapi manages the system timezone of target hosts with `timedatectl`. The agent
reads and sets the timezone atomically. The update operation is idempotent. If
the timezone already has the requested value, the agent changes nothing and
returns `changed: false`.

## How it works

The timezone provider uses `timedatectl` to read and set the system timezone.
The get operation reads the current timezone name and the UTC offset. The update
operation compares the current timezone with the requested value. It calls
`timedatectl set-timezone` only when a change is necessary.

### Get

Runs `timedatectl show -p Timezone --value` to read the IANA timezone name, for
example `America/New_York`, and `date +%:z` to read the current UTC offset, for
example `-05:00`.

### Update

Compares the current timezone against the requested value. If they match,
returns `changed: false`. Otherwise, runs `timedatectl set-timezone <timezone>`
and returns `changed: true`.

## Operations

| Operation | Description                                                   |
| --------- | ------------------------------------------------------------- |
| Get       | Get the current timezone name and UTC offset                  |
| Update    | Set the timezone (idempotent, returns changed: false if same) |

## CLI usage

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

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All timezone operations accept broadcast targets. Use `--target _all` to query
or update all registered agents. Use a label selector, for example
`--target group:web`, to target a part of the fleet.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  TIMEZONE          UTC_OFFSET
  web-01    America/New_York  -05:00
  web-02    UTC               +00:00
```

Skipped and failed hosts appear with their error in the output.

## Idempotent updates

The update operation compares the requested timezone with the current value. If
they are the same, the response is `changed: false`, and the agent makes no
system call. You can run this operation many times safely.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, timezone operations return
`status: skipped`. They do not fail. Refer to
[Platform Detection](../sdk/platform/detection.md) for how osapi finds the OS
family.

## Permissions

| Operation | Permission       |
| --------- | ---------------- |
| Get       | `timezone:read`  |
| Update    | `timezone:write` |

All built-in roles (`admin`, `write`, `read`) include `timezone:read`. The
`admin` and `write` roles also include `timezone:write`.

## Related

- [NTP Management](ntp-management.md): NTP server configuration
- [Configuration](../usage/configuration.md): full configuration reference
