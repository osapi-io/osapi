# Create

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
