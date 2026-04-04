# Get

Get the OSAPI-managed static routes for a specific interface:

```bash
$ osapi client node network route get \
    --target web-01 --interface eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  DESTINATION   GATEWAY      INTERFACE  METRIC
  10.0.0.0/8    192.168.1.1  eth0       0
  172.16.0.0/12 192.168.1.1  eth0       100
```

When targeting all hosts:

```bash
$ osapi client node network route get \
    --target _all --interface eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  DESTINATION   GATEWAY      INTERFACE  METRIC
  10.0.0.0/8    192.168.1.1  eth0       0

  web-02
  DESTINATION   GATEWAY      INTERFACE  METRIC
  10.0.0.0/8    192.168.1.1  eth0       0
```

When some hosts fail or are skipped, STATUS and ERROR columns are shown:

```bash
$ osapi client node network route get \
    --target _all --interface eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   DESTINATION   GATEWAY  INTERFACE  METRIC  ERROR
  web-01    ok       10.0.0.0/8    ...
  mac-01    skipped                                             unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network route get \
    --target web-01 --interface eth0 --json
{"results":[{"hostname":"web-01","status":"ok","routes":[
{"destination":"10.0.0.0/8","gateway":"192.168.1.1",
"interface":"eth0","metric":0}
]}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--interface`  | Interface name                                           | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
