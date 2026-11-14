# Update

Replace a Netplan interface configuration on a target host. The command fails if
osapi does not manage a configuration for that interface. Use `create` first. If
the content did not change, the command returns `changed: false`:

```bash
$ osapi client node network interface update \
    --target web-01 --name eth0 \
    --address 192.168.1.200/24 \
    --gateway4 192.168.1.1 \
    --mtu 9000

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED
  web-01    changed  eth0    true

  1 host: 1 changed
```

When the configuration does not change, `changed` is false:

```bash
$ osapi client node network interface update \
    --target web-01 --name eth0 \
    --address 192.168.1.200/24 \
    --gateway4 192.168.1.1 \
    --mtu 9000

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME    CHANGED
  web-01    ok      eth0    false

  1 host: 1 ok
```

Broadcast an update to all hosts:

```bash
$ osapi client node network interface update \
    --target _all --name eth0 --dhcp4

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED
  web-01    changed  eth0    true
  web-02    changed  eth0    true

  2 hosts: 2 changed
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network interface update \
    --target web-01 --name eth0 \
    --address 192.168.1.200/24 --json
{"results":[{"hostname":"web-01","name":"eth0",
"changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag            | Description                                              | Default  |
| --------------- | -------------------------------------------------------- | -------- |
| `--name`        | Interface name                                           | required |
| `--dhcp4`       | Enable DHCPv4                                            |          |
| `--dhcp6`       | Enable DHCPv6                                            |          |
| `--address`     | IP address in CIDR notation (repeatable)                 |          |
| `--gateway4`    | IPv4 gateway address                                     |          |
| `--gateway6`    | IPv6 gateway address                                     |          |
| `--mtu`         | Maximum transmission unit                                |          |
| `--mac-address` | Hardware MAC address                                     |          |
| `--wakeonlan`   | Enable Wake-on-LAN                                       |          |
| `-T, --target`  | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`    | Output raw JSON response                                 |          |
