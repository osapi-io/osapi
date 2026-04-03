# List

List all routes in the kernel routing table on a target host:

```bash
$ osapi client node network route list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  DESTINATION     GATEWAY        INTERFACE  METRIC
  0.0.0.0/0       192.168.1.1    eth0       100
  10.0.0.0/8      192.168.1.1    eth0       0
  192.168.1.0/24  0.0.0.0        eth0       0
```

Target all hosts to list routes across the fleet:

```bash
$ osapi client node network route list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  DESTINATION     GATEWAY        INTERFACE  METRIC
  0.0.0.0/0       192.168.1.1    eth0       100
  192.168.1.0/24  0.0.0.0        eth0       0

  web-02
  DESTINATION     GATEWAY        INTERFACE  METRIC
  0.0.0.0/0       10.0.0.1       eth0       100
  10.0.0.0/24     0.0.0.0        eth0       0
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network route list --target web-01 --json
{"results":[{"hostname":"web-01","status":"ok","routes":[
{"destination":"0.0.0.0/0","gateway":"192.168.1.1",
"interface":"eth0","metric":100}
]}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
