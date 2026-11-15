# Update

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
returns `Changed: false`. To change the values of a template, give each
`--var key=value` again.
