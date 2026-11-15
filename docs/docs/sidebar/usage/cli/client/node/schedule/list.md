# List

List all osapi-managed scheduled entries:

```bash
$ osapi client node schedule list --target web-01

  HOSTNAME  STATUS  NAME           SCHEDULE     OBJECT          USER
  web-01    ok      backup-daily   0 2 * * *    backup-script   root
  web-01    ok      log-rotate     0 0 * * 0    logrotate-conf  root

  1 host: 1 ok
```
