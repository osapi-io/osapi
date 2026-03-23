---
sidebar_position: 9
---

# Cron Management

OSAPI manages cron entries on target hosts. It supports two placement modes:

- **Custom schedule** — writes to `/etc/cron.d/{name}` with a 5-field cron
  expression
- **Periodic interval** — writes to `/etc/cron.{hourly,daily,weekly,monthly}/`
  as executable scripts

Cron entries reference scripts stored in the NATS Object Store by name.
Upload a script first with the file management commands, then create a cron
entry pointing at it. This separates script content from scheduling
configuration and enables versioned updates.

## How It Works

### Object-Based Workflow

The cron provider is a **meta provider**: it does not embed script content
directly. Instead, each cron entry holds an `object` name that references a
file in the NATS Object Store. When the agent deploys or updates a cron entry,
it:

1. Fetches the named object from the Object Store.
2. Writes the script content to the appropriate path under `/etc/cron.d/` or
   `/etc/cron.{interval}/`.
3. Records the deploy state in the file-state KV bucket (SHA-256, path,
   timestamps).

The file-state KV tracks each deployed cron file, so updates are idempotent:
if the script content has not changed, the file is not rewritten.

### Custom Schedule (`/etc/cron.d/`)

For entries with a `schedule` field, the provider writes a drop-in file. The
file content comes from the referenced object in the Object Store:

```
0 2 * * * root /usr/local/bin/backup.sh --full
```

The file is owned by root with mode `0644`. No header comment is written;
ownership is tracked via the file-state KV instead.

### Periodic Interval (`/etc/cron.{interval}/`)

For entries with an `interval` field, the provider writes an executable script
to the matching interval directory. The file content comes from the referenced
object — it should include a `#!/bin/sh` shebang:

```bash
#!/bin/sh
/usr/sbin/logrotate /etc/logrotate.conf
```

The file is written with mode `0755` so `run-parts` can execute it.

`schedule` and `interval` are mutually exclusive — provide exactly one. The
API validates this and returns 400 if both or neither are provided.

### File-State KV Tracking

Deploy state is recorded in the file-state KV bucket rather than embedded in
the deployed file as a header comment. This means:

- Deployed cron files contain only the script content — no `# Managed by
  osapi` marker.
- The list and get operations query the file-state KV to discover managed
  entries; manually created files are left untouched.
- State persists in the KV until explicitly removed — deleting a cron entry
  undeploys the file from disk but preserves the file-state record.

### Template Support

If the referenced object was uploaded with `content_type: template`, the agent
renders it as a Go `text/template` before writing to disk. Pass variables at
deploy time with `--vars`:

```bash
# Upload a template script
osapi client node file upload --name backup-script.tmpl \
  --content-type template --file ./backup.sh.tmpl

# Create a cron entry that renders the template on deploy
osapi client node schedule cron create --target web-01 \
  --name backup --schedule "0 2 * * *" \
  --object backup-script.tmpl \
  --vars "retention_days=30,s3_bucket=my-bucket"
```

Template variables are merged with the agent's system facts and hostname. See
[File Management](file-management.md) for the full template context reference.

## Operations

| Operation | Description                              |
| --------- | ---------------------------------------- |
| List      | List all osapi-managed cron entries      |
| Get       | Get a specific entry by name             |
| Create    | Upload script, then create cron entry    |
| Update    | Upload new script version, then update   |
| Delete    | Undeploy cron file from disk             |

## CLI Usage

```bash
# Upload a script to the Object Store first
osapi client node file upload --name backup-script \
  --file ./backup.sh

# Create with a custom schedule (/etc/cron.d/)
osapi client node schedule cron create --target web-01 \
  --name backup --schedule "0 2 * * *" \
  --object backup-script --user root

# Create with an interval (/etc/cron.daily/)
osapi client node schedule cron create --target web-01 \
  --name logrotate --interval daily \
  --object logrotate-script

# List all managed cron entries
osapi client node schedule cron list --target web-01

# Get a specific entry
osapi client node schedule cron get --target web-01 --name backup

# Update: upload a new script version and redeploy
osapi client node file upload --name backup-script \
  --file ./backup-v2.sh --force
osapi client node schedule cron update --target web-01 \
  --name backup --schedule "0 3 * * *" \
  --object backup-script

# Delete an entry (undeploys file from disk; state preserved in KV)
osapi client node schedule cron delete --target web-01 --name backup
```

All commands support `--json` for raw JSON output.

## File Permissions

The provider sets file ownership and modes to match cron requirements:

| Type     | Path                          | Mode | Notes                           |
| -------- | ----------------------------- | ---- | ------------------------------- |
| Schedule | `/etc/cron.d/{name}`          | 0644 | Not executable, root-owned      |
| Interval | `/etc/cron.{interval}/{name}` | 0755 | Executable, `#!/bin/sh` shebang |

Files are created as root (the agent runs as root). Names must not contain dots
— `run-parts` skips dotfiles.

## Undeploy Behavior

Deleting a cron entry **undeploys** the file: it is removed from the
filesystem, but the file-state KV record is preserved. This means:

- Re-creating the entry with the same name and object will detect the prior
  state and only write the file if the content differs.
- The KV record serves as an audit trail of what was last deployed.

To remove the file-state record entirely, delete the corresponding file-state
entry via the file management API after removing the cron entry.

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
`^[a-zA-Z0-9_-]+$`). Names like `backup-daily`, `log_rotate`, and
`health_check` are valid. Names containing `/`, `..`, or spaces are rejected.

## Idempotent Updates

The update operation compares the new content against the existing file using
the SHA-256 stored in the file-state KV. If the script content and scheduling
parameters are unchanged, the response returns `changed: false`. This makes it
safe to run updates repeatedly without generating unnecessary filesystem writes.

## Future: Crontab and Systemd Timers

Cron drop-in management is the first provider under the `schedule` domain.
Future providers will add:

- **Crontab** — user-level crontab management (`crontab -u <user>`)
- **Systemd Timers** — systemd timer unit management

These will be separate API endpoints under `/node/{hostname}/schedule/crontab`
and `/node/{hostname}/schedule/timer`.

## Related

- [File Management](file-management.md) — uploading scripts and template
  rendering
- [CLI Reference](../usage/cli/client/node/schedule/cron.md) — cron commands
- [Platform Detection](../sdk/platform.md) — OS family detection
- [Configuration](../usage/configuration.md) — full configuration reference
