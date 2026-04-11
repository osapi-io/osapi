# Create

Create static routes for an interface on a target host. Routes are specified in
`TO:VIA` or `TO:VIA:METRIC` format. Fails if an OSAPI-managed routes
configuration for that interface already exists -- use `update` to replace it:

```bash
$ osapi client node network route create \
    --target web-01 --interface eth0 \
    --route 10.0.0.0/8:192.168.1.1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   INTERFACE  CHANGED
  web-01    changed  eth0       true

  1 host: 1 changed
```

Create multiple routes at once:

```bash
$ osapi client node network route create \
    --target web-01 --interface eth0 \
    --route 10.0.0.0/8:192.168.1.1 \
    --route 172.16.0.0/12:192.168.1.1:100

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   INTERFACE  CHANGED
  web-01    changed  eth0       true

  1 host: 1 changed
```

Broadcast to all hosts at once:

```bash
$ osapi client node network route create \
    --target _all --interface eth0 \
    --route 10.0.0.0/8:192.168.1.1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   INTERFACE  CHANGED
  web-01    changed  eth0       true
  web-02    changed  eth0       true

  2 hosts: 2 changed
```

When some hosts are skipped:

```bash
  HOSTNAME  STATUS   INTERFACE  CHANGED
  web-01    changed  eth0       true
  mac-01    skip

  2 hosts: 1 changed, 1 skipped

  Details:
  mac-01    unsupported platform
```

## Route Format

Routes are specified as `TO:VIA` or `TO:VIA:METRIC`:

| Format          | Example                      |
| --------------- | ---------------------------- |
| `TO:VIA`        | `10.0.0.0/8:192.168.1.1`     |
| `TO:VIA:METRIC` | `10.0.0.0/8:192.168.1.1:100` |

`TO` is the destination in CIDR notation. `VIA` is the gateway IP address.
`METRIC` is the route priority (lower is preferred).

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network route create \
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
