# Update

Replace the static routes for an interface on a target host. Fails if no
OSAPI-managed routes configuration exists for that interface -- use `create`
first. Returns `changed: false` if the content has not changed:

```bash
$ osapi client node network route update \
    --target web-01 --interface eth0 \
    --route 10.0.0.0/8:192.168.1.1 \
    --route 172.16.0.0/12:192.168.1.1:100

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  INTERFACE  CHANGED
  eth0       true
```

When the configuration is unchanged, `changed` is false:

```bash
$ osapi client node network route update \
    --target web-01 --interface eth0 \
    --route 10.0.0.0/8:192.168.1.1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  INTERFACE  CHANGED
  eth0       false
```

Broadcast an update to all hosts:

```bash
$ osapi client node network route update \
    --target _all --interface eth0 \
    --route 10.0.0.0/8:10.0.0.1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  INTERFACE  CHANGED
  web-01    eth0       true
  web-02    eth0       true
```

## Route Format

Routes are specified as `TO:VIA` or `TO:VIA:METRIC`:

| Format          | Example                      |
| --------------- | ---------------------------- |
| `TO:VIA`        | `10.0.0.0/8:192.168.1.1`     |
| `TO:VIA:METRIC` | `10.0.0.0/8:192.168.1.1:100` |

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network route update \
    --target web-01 --interface eth0 \
    --route 10.0.0.0/8:192.168.1.1 --json
{"results":[{"hostname":"web-01","interface":"eth0",
"changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--interface`  | Interface name                                           | required |
| `--route`      | Route in `TO:VIA` or `TO:VIA:METRIC` format (repeatable) | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
