# Update

Set the hostname on the target node:

```bash
$ osapi client node hostname update --name web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

When the target host does not support the operation:

```bash
$ osapi client node hostname update --name web-01 --target darwin-host

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME     STATUS
  darwin-host  skip

  1 host: 1 skipped

  Details:
  darwin-host  host: operation not supported on this OS family
```

When targeting all hosts:

```bash
$ osapi client node hostname update --name web-01 --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  server1   changed  true
  server2   skip

  2 hosts: 1 changed, 1 skipped

  Details:
  server2   unsupported platform
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | New hostname to set (required)                           | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
