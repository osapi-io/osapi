---
sidebar_position: 1
---

# Schedule

Manage scheduled entries on target hosts. On Debian family hosts, these are cron
drop-in files in `/etc/cron.d/`.

## List

List all osapi-managed scheduled entries:

```bash
$ osapi client node schedule list --target web-01

  HOSTNAME  STATUS  NAME           SCHEDULE     OBJECT          USER
  web-01    ok      backup-daily   0 2 * * *    backup-script   root
  web-01    ok      log-rotate     0 0 * * 0    logrotate-conf  root

  1 host: 1 ok
```

## Get

Get a specific scheduled entry by name:

```bash
$ osapi client node schedule get --target web-01 --name backup-daily

  HOSTNAME  STATUS  NAME           SCHEDULE     OBJECT          USER
  web-01    ok      backup-daily   0 2 * * *    backup-script   root

  1 host: 1 ok
```

## Create

Upload the script to the Object Store first. Then make the scheduled entry, and
refer to the script by its object name:

```bash
$ osapi client file upload --name backup-script \
    --file /usr/local/bin/backup.sh
```

Then make the entry, with `--object` to refer to the uploaded file:

```bash
$ osapi client node schedule create --target web-01 \
    --name backup-daily \
    --schedule "0 2 * * *" \
    --object backup-script \
    --user root

  HOSTNAME  STATUS   NAME          CHANGED
  web-01    changed  backup-daily  true

  1 host: 1 changed
```

If you do not give `--user`, the value is `root`.

If the object is a Go template, use `--content-type template`. Then the agent
renders it with its facts before it writes the file to disk.

## Update

Update an existing scheduled entry:

```bash
$ osapi client node schedule update --target web-01 \
    --name backup-daily \
    --schedule "0 3 * * *"

  HOSTNAME  STATUS   NAME          CHANGED
  web-01    changed  backup-daily  true

  1 host: 1 changed
```

The command updates only the fields that you give. If nothing changed, it
returns `Changed: false`.

## Delete

Delete a scheduled entry:

```bash
$ osapi client node schedule delete --target web-01 --name backup-daily

  HOSTNAME  STATUS   NAME          CHANGED
  web-01    changed  backup-daily  true

  1 host: 1 changed
```

## JSON Output

All commands accept `--json` to give the raw JSON output:

```bash
$ osapi client node schedule list --target web-01 --json
{"results":[{"name":"backup-daily","schedule":"0 2 * * *","user":"root","object":"backup-script"}],"job_id":"..."}
```
