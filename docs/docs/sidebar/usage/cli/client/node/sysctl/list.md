# List

List all OSAPI-managed sysctl parameters on a target host:

```bash
$ osapi client node sysctl list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  KEY                      VALUE
  web-01    ok      net.ipv4.ip_forward      1
  web-01    ok      vm.swappiness            10
  web-01    ok      kernel.panic             30

  1 host: 1 ok
```

Target all hosts to list parameters across the fleet:

```bash
$ osapi client node sysctl list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  KEY                  VALUE
  web-01    ok      net.ipv4.ip_forward  1
  web-01    ok      vm.swappiness        10
  web-02    ok      net.ipv4.ip_forward  1
  web-02    ok      vm.swappiness        10

  2 hosts: 2 ok
```

Target by label to list parameters on a group of servers:

```bash
$ osapi client node sysctl list --target group:web
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node sysctl list --target web-01 --json
{"results":[{"hostname":"web-01","key":"net.ipv4.ip_forward","value":"1","status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
