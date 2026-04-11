# Get

Get a specific sysctl parameter by key:

```bash
$ osapi client node sysctl get --target web-01 --key net.ipv4.ip_forward

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  KEY                   VALUE
  web-01    ok      net.ipv4.ip_forward   1

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node sysctl get --target _all --key vm.swappiness

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  KEY            VALUE
  web-01    ok      vm.swappiness  10
  web-02    ok      vm.swappiness  10

  2 hosts: 2 ok
```

When some hosts are skipped:

```bash
$ osapi client node sysctl get --target _all --key vm.swappiness

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  KEY            VALUE
  web-01    ok      vm.swappiness  10
  mac-01    skip

  2 hosts: 1 ok, 1 skipped

  Details:
  mac-01    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node sysctl get --target web-01 --key vm.swappiness --json
{"results":[{"hostname":"web-01","key":"vm.swappiness","value":"10",
"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--key`        | Sysctl parameter key to retrieve                         | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
