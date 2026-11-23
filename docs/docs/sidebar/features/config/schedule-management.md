# Schedule management

osapi manages scheduled entries on target hosts. It has two placement modes:

- **Custom schedule.** The provider writes `/etc/cron.d/{name}` with a 5-field
  cron expression.
- **Periodic interval.** The provider writes an executable script to
  `/etc/cron.{hourly,daily,weekly,monthly}/`.

A scheduled entry refers by name to a script in the NATS Object Store. First,
upload a script with the file management commands. Then make an entry that
points to it. Thus the script content and the schedule configuration are
separate, and you can update a script as a new version.

The `schedule` behavior supplies the API, the CLI commands, the SDK service and
the UI blocks. The
[behavior reference](../../reference/behaviors/config/schedule/schedule.md) has
each method, its fields and its examples.

## How it works

### Object-based workflow

The behavior does not keep the script content. Each entry has an `object` name
that refers to a file in the NATS Object Store. When the agent deploys or
updates an entry, it does these steps:

1. It gets the named object from the Object Store, through the `object` building
   block.
1. It writes the script content to the correct path under `/etc/cron.d/` or
   `/etc/cron.{interval}/`, through the `managedfile` building block.
1. `managedfile` records the deploy state in the file-state KV bucket (SHA-256,
   path, times, and the object, schedule, interval and user of the entry).

`create` and `update` are idempotent on the content. The agent compares the
requested content with the file on disk, and writes the file only when they are
different. Thus it puts back an entry that somebody changed by hand.

`create` has two exceptions:

- If osapi deployed an entry with the same name in a different directory,
  `create` reports `changed: false`. It does not write a second copy. To move
  the entry, use `delete`, then `create`.
- If a file is at the path and osapi did not deploy it, `create` refuses the
  request. osapi does not write over a cron file that somebody made by hand.

### Custom schedule (`/etc/cron.d/`)

For an entry with a `schedule` field, the provider writes a drop-in file. The
file content comes from the object in the Object Store:

```
0 2 * * * root /usr/local/bin/backup.sh --full
```

root owns the file, and its mode is `0644`. The provider writes no header
comment. The file-state KV records the ownership.

### Periodic interval (`/etc/cron.{interval}/`)

For an entry with an `interval` field, the provider writes an executable script
to the directory of that interval. The file content comes from the object. It
must start with a `#!/bin/sh` line:

```bash
#!/bin/sh
/usr/sbin/logrotate /etc/logrotate.conf
```

The provider writes the file with mode `0755`, so that `run-parts` can run it.

Give `schedule` or `interval`, not both. The API checks this. If you give both
or neither, it returns 400.

### File-state KV tracking

The provider records the deploy state in the file-state KV bucket. It does not
put a header comment in the deployed file. Thus:

- A deployed cron file has only the script content. It has no
  `# Managed by osapi` marker.
- The list and get operations read the file-state KV to find the managed
  entries. They do not touch files that somebody made by hand.
- The state stays in the KV until you remove it. When you delete an entry, the
  agent removes the file from disk and marks the file-state record as
  undeployed.

### Template support

If the object has `content_type: template`, the agent renders it as a Go
`text/template` before it writes the file to disk. To render an object of a
different content type, give `--content-type template`. Give the variables with
`--vars`, as `key=value` pairs:

```bash
# Upload a template script
osapi client file upload --name backup-script.tmpl \
  --content-type template --file ./backup.sh.tmpl

# Create a scheduled entry that renders the template on deploy
osapi client node schedule create --target web-01 \
  --name backup --schedule "0 2 * * *" \
  --object backup-script.tmpl \
  --vars retention_days=30,s3_bucket=my-bucket
```

The template reads the variables as `.Vars`, the facts that the agent collected
last as `.Facts`, and the hostname as `.Hostname`. Refer to
[File Management](file-management.md) for the full template context.

## Operations

| Operation | Description                                   |
| --------- | --------------------------------------------- |
| List      | List all entries that osapi manages           |
| Get       | Get one entry by name                         |
| Create    | Upload the script, then make the entry        |
| Update    | Deploy the entry again, with the fields given |
| Delete    | Remove the cron file from disk                |

## CLI usage

```bash
# Upload a script to the Object Store first
osapi client file upload --name backup-script \
  --file ./backup.sh

# Create with a custom schedule (/etc/cron.d/)
osapi client node schedule create --target web-01 \
  --name backup --schedule "0 2 * * *" \
  --object backup-script --user root

# Create with an interval (/etc/cron.daily/)
osapi client node schedule create --target web-01 \
  --name logrotate --interval daily \
  --object logrotate-script

# List all managed entries
osapi client node schedule list --target web-01

# Get a specific entry
osapi client node schedule get --target web-01 --name backup

# Update: upload a new script version and redeploy
osapi client file upload --name backup-script-v2 \
  --file ./backup-v2.sh
osapi client node schedule update --target web-01 \
  --name backup --schedule "0 3 * * *" \
  --object backup-script-v2

# Delete an entry (removes the file; the KV record stays, marked undeployed)
osapi client node schedule delete --target web-01 --name backup
```

All commands accept `--json` to give the raw JSON output.

## File permissions

The provider sets the file owner and mode that cron needs:

| Type     | Path                          | Mode | Notes                               |
| -------- | ----------------------------- | ---- | ----------------------------------- |
| Schedule | `/etc/cron.d/{name}`          | 0644 | Not executable, root is the owner   |
| Interval | `/etc/cron.{interval}/{name}` | 0755 | Executable, starts with `#!/bin/sh` |

The agent writes the files as root. A name must not have dots, because
`run-parts` does not run files with dots in their names.

## Undeploy behavior

When you delete an entry, the agent **undeploys** the file. It removes the file
from the filesystem, and marks the file-state KV record as undeployed. Thus:

- When you make the entry again with the same name and object, the agent finds
  the previous state. It writes the file only if the content is different.
- The KV record is an audit trail of the last deploy.

To remove the file-state record, delete the file-state entry through the file
management API after you remove the scheduled entry.

## Supported platforms

| OS family | List, Get | Create, Update, Delete |
| --------- | --------- | ---------------------- |
| Debian    | Full      | Full                   |
| Darwin    | Skipped   | Skipped                |

On a platform that is not supported, schedule operations return
`status: skipped`. They do not fail. Refer to
[Platform Detection](../../sdk/platform/detection.md) for how osapi finds the OS
family.

## Permissions

The [schedule](../../reference/behaviors/config/schedule/schedule.md) reference
page gives the permission of each method.

## Naming rules

An entry name can have only letters, digits, hyphens and underscores (pattern:
`^[a-zA-Z0-9_-]+$`). For example, `backup-daily`, `log_rotate` and
`health_check` are valid. The API rejects names with `/`, `..` or spaces.

## Idempotent updates

The update operation compares the new content with the file that exists, and the
object, schedule and user with the record of the entry. If none of them changed,
the response is `changed: false`. A new schedule or user with the same content
changes only the record, so that list and get give the new values. Thus you can
run an update many times, and it writes to the filesystem only when necessary.
An update keeps the entry in its directory: it cannot change the interval. A
field that the update does not give keeps the value of the entry.

## Future: crontab and systemd timers

Cron drop-in management is the first part of the `schedule` behavior. Future
work will add:

- **Crontab**: user crontab management (`crontab -u <user>`)
- **Systemd timers**: systemd timer unit management

Each will have its own routes.

## Related

- [File Management](file-management.md): upload scripts and render templates
- [CLI Reference](../../usage/cli/client/node/schedule/schedule.mdx): the
  schedule commands
- [Platform Detection](../../sdk/platform/detection.md): how osapi finds the OS
  family
- [Configuration](../../usage/configuration.md): all configuration settings
