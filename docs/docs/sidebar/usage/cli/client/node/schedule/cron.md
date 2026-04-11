---
sidebar_position: 1
---

# Cron

Manage cron drop-in files in `/etc/cron.d/` on target hosts.

## List

List all osapi-managed cron entries:

```bash
$ osapi client node schedule cron list --target web-01

  HOSTNAME  STATUS  NAME           SCHEDULE     OBJECT          USER
  web-01    ok      backup-daily   0 2 * * *    backup-script   root
  web-01    ok      log-rotate     0 0 * * 0    logrotate-conf  root

  1 host: 1 ok
```

## Get

Get a specific cron entry by name:

```bash
$ osapi client node schedule cron get --target web-01 --name backup-daily

  HOSTNAME  STATUS  NAME           SCHEDULE     OBJECT          USER
  web-01    ok      backup-daily   0 2 * * *    backup-script   root

  1 host: 1 ok
```

## Create

Upload the script to the Object Store first, then create the cron entry
referencing it by object name:

```bash
$ osapi client file upload --name backup-script \
    --file /usr/local/bin/backup.sh
```

Then create the cron entry using `--object` to reference the uploaded file:

```bash
$ osapi client node schedule cron create --target web-01 \
    --name backup-daily \
    --schedule "0 2 * * *" \
    --object backup-script \
    --user root

  HOSTNAME  STATUS   NAME          CHANGED
  web-01    changed  backup-daily  true

  1 host: 1 changed
```

The `--user` flag defaults to `root` if omitted.

Use `--content-type template` if the object was uploaded as a Go template and
should be rendered with agent facts before being written to disk.

## Update

Update an existing cron entry:

```bash
$ osapi client node schedule cron update --target web-01 \
    --name backup-daily \
    --schedule "0 3 * * *"

  HOSTNAME  STATUS   NAME          CHANGED
  web-01    changed  backup-daily  true

  1 host: 1 changed
```

Only the fields you specify are updated. If nothing changed, `Changed: false`.

## Delete

Delete a cron entry:

```bash
$ osapi client node schedule cron delete --target web-01 --name backup-daily

  HOSTNAME  STATUS   NAME          CHANGED
  web-01    changed  backup-daily  true

  1 host: 1 changed
```

## JSON Output

All commands support `--json` for raw JSON output:

```bash
$ osapi client node schedule cron list --target web-01 --json
{"results":[{"name":"backup-daily","schedule":"0 2 * * *","user":"root","object":"backup-script"}],"job_id":"..."}
```
