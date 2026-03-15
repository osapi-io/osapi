# Status

Show health status for all registered API servers:

```bash
$ osapi client api status

  TYPE  HOSTNAME       STATUS  CONDITIONS  AGE    CPU    MEM
  api   api-server-01  Ready   -           7h 6m  2.1%   128 MB
```

Use `--json` for the full health status JSON response:

```bash
$ osapi client api status --json
```
