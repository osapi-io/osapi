---
sidebar_position: 12
---

# Timezone management

osapi reads the system timezone of each host, and sets it on the Debian family.
It is the [timezone behavior](../behaviors/timezone/timezone.md). The update is
idempotent. If the host already has the timezone, the agent changes nothing and
returns `changed: false`.

## How it works

### Get

gohai reads the timezone on every Linux family and on macOS. The name is the
zone that `/etc/localtime` links to, such as `America/New_York`. On Linux, when
there is no link, the name is in `/etc/timezone`. The offset is the offset of
the host's clock from UTC now, such as `-05:00`.

### Update

Runs `timedatectl show -p Timezone --value` to read the current timezone. If it
is the requested value, the update returns `changed: false`. Otherwise it runs
`timedatectl set-timezone <timezone>` and returns `changed: true`. The update
runs on the Debian family only. On other hosts it reports skipped. The name must
be an IANA name, such as `UTC` or `America/Argentina/Buenos_Aires`.

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
