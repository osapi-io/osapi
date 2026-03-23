---
sidebar_position: 1
---

# Cron

Manage cron drop-in files in `/etc/cron.d/` on target hosts.

## List

List all osapi-managed cron entries:

```bash
$ osapi client node schedule cron list --target web-01

  NAME           SCHEDULE     USER  COMMAND
  backup-daily   0 2 * * *    root  /usr/local/bin/backup.sh
  log-rotate     0 0 * * 0    root  /usr/sbin/logrotate /etc/logrotate.conf
```

## Get

Get a specific cron entry by name:

```bash
$ osapi client node schedule cron get --target web-01 --name backup-daily

  Name: backup-daily
  Schedule: 0 2 * * *
  User: root
  Command: /usr/local/bin/backup.sh
```

## Create

Create a new cron entry:

```bash
$ osapi client node schedule cron create --target web-01 \
    --name backup-daily \
    --schedule "0 2 * * *" \
    --command "/usr/local/bin/backup.sh" \
    --user root

  Name: backup-daily
  Changed: true
```

The `--user` flag defaults to `root` if omitted.

## Update

Update an existing cron entry:

```bash
$ osapi client node schedule cron update --target web-01 \
    --name backup-daily \
    --schedule "0 3 * * *"

  Name: backup-daily
  Changed: true
```

Only the fields you specify are updated. If nothing changed, `Changed: false`.

## Delete

Delete a cron entry:

```bash
$ osapi client node schedule cron delete --target web-01 --name backup-daily

  Name: backup-daily
  Changed: true
```

## JSON Output

All commands support `--json` for raw JSON output:

```bash
$ osapi client node schedule cron list --target web-01 --json
{"results":[{"name":"backup-daily","schedule":"0 2 * * *","user":"root","command":"/usr/local/bin/backup.sh"}],"job_id":"..."}
```
