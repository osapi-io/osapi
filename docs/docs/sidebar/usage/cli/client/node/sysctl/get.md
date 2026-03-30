# Get

Get a specific sysctl parameter by key:

```bash
$ osapi client node sysctl get --target web-01 --key net.ipv4.ip_forward

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  KEY                   VALUE
  net.ipv4.ip_forward   1
```

When targeting all hosts:

```bash
$ osapi client node sysctl get --target _all --key vm.swappiness

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  KEY           VALUE
  web-01    vm.swappiness  10
  web-02    vm.swappiness  10
```

When some hosts are skipped, a STATUS column is shown:

```bash
$ osapi client node sysctl get --target _all --key vm.swappiness

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   KEY           VALUE  ERROR
  web-01    ok       vm.swappiness  10
  mac-01    skipped                        unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node sysctl get --target web-01 --key vm.swappiness --json
{"results":[{"hostname":"web-01","key":"vm.swappiness","value":"10","status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--key`        | Sysctl parameter key to retrieve                         | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
