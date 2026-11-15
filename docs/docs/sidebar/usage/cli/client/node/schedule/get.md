# Get

Get a specific scheduled entry by name:

```bash
$ osapi client node schedule get --target web-01 --name backup-daily

  HOSTNAME  STATUS  NAME           SCHEDULE     OBJECT          USER
  web-01    ok      backup-daily   0 2 * * *    backup-script   root

  1 host: 1 ok
```
