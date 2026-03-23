---
sidebar_position: 9
---

# Cron Management

OSAPI manages cron entries on target hosts. It supports two placement modes:

- **Custom schedule** — writes to `/etc/cron.d/{name}` with a 5-field cron
  expression
- **Periodic interval** — writes to `/etc/cron.{hourly,daily,weekly,monthly}/`
  as executable scripts

## How It Works

### Custom Schedule (`/etc/cron.d/`)

For entries with a `schedule` field, the provider writes a drop-in file:

```
# Managed by osapi
0 2 * * * root /usr/local/bin/backup.sh
```

### Periodic Interval (`/etc/cron.{interval}/`)

For entries with an `interval` field, the provider writes an executable script:

```bash
#!/bin/sh
# Managed by osapi
/usr/sbin/logrotate /etc/logrotate.conf
```

`schedule` and `interval` are mutually exclusive — provide exactly one. The API
validates this and returns 400 if both or neither are provided.

Only files with the `# Managed by osapi` header are listed and managed. Manually
created files are left untouched.

## Operations

| Operation | Description                         |
| --------- | ----------------------------------- |
| List      | List all osapi-managed cron entries |
| Get       | Get a specific entry by name        |
| Create    | Create a new cron drop-in file      |
| Update    | Update an existing entry            |
| Delete    | Remove a cron drop-in file          |

## CLI Usage

```bash
# List all managed cron entries (from /etc/cron.d/ and /etc/cron.{daily,weekly,...}/)
osapi client node schedule cron list --target web-01

# Get a specific entry
osapi client node schedule cron get --target web-01 --name backup

# Create with a custom schedule (/etc/cron.d/)
osapi client node schedule cron create --target web-01 \
  --name backup --schedule "0 2 * * *" \
  --command "/usr/local/bin/backup.sh" --user root

# Create with an interval (/etc/cron.daily/)
osapi client node schedule cron create --target web-01 \
  --name logrotate --interval daily \
  --command "/usr/sbin/logrotate /etc/logrotate.conf"

# Update the schedule
osapi client node schedule cron update --target web-01 \
  --name backup --schedule "0 3 * * *"

# Delete an entry
osapi client node schedule cron delete --target web-01 --name backup
```

All commands support `--json` for raw JSON output.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |

On unsupported platforms, cron operations return `status: skipped` instead of
failing. See [Platform Detection](../sdk/platform.md) for details on OS family
detection.

## Permissions

| Operation              | Permission   |
| ---------------------- | ------------ |
| List, Get              | `cron:read`  |
| Create, Update, Delete | `cron:write` |

All built-in roles (`admin`, `write`, `read`) include `cron:read`. The `admin`
and `write` roles also include `cron:write`.

## Naming Rules

Entry names must be alphanumeric with hyphens and underscores only (pattern:
`^[a-zA-Z0-9_-]+$`). Names like `backup-daily`, `log_rotate`, and `health_check`
are valid. Names containing `/`, `..`, or spaces are rejected.

## Idempotent Updates

The update operation compares the new content against the existing file. If
nothing changed, the response returns `changed: false`. This makes it safe to
run updates repeatedly without generating unnecessary changes.

## Future: Crontab and Systemd Timers

Cron drop-in management is the first provider under the `schedule` domain.
Future providers will add:

- **Crontab** — user-level crontab management (`crontab -u <user>`)
- **Systemd Timers** — systemd timer unit management

These will be separate API endpoints under `/node/{hostname}/schedule/crontab`
and `/node/{hostname}/schedule/timer`.

## Related

- [CLI Reference](../usage/cli/client/node/schedule/cron.md) — cron commands
- [Platform Detection](../sdk/platform.md) — OS family detection
- [Configuration](../usage/configuration.md) — full configuration reference
