# Update

Replace an existing Netplan interface configuration on a target host. Fails if
no OSAPI-managed configuration exists for that interface -- use `create` first.
Returns `changed: false` if the content has not changed:

```bash
$ osapi client node network interface update \
    --target web-01 --name eth0 \
    --address 192.168.1.200/24 \
    --gateway4 192.168.1.1 \
    --mtu 9000

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME    CHANGED
  eth0    true
```

When the configuration is unchanged, `changed` is false:

```bash
$ osapi client node network interface update \
    --target web-01 --name eth0 \
    --address 192.168.1.200/24 \
    --gateway4 192.168.1.1 \
    --mtu 9000

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME    CHANGED
  eth0    false
```

Broadcast an update to all hosts:

```bash
$ osapi client node network interface update \
    --target _all --name eth0 --dhcp4

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    CHANGED
  web-01    eth0    true
  web-02    eth0    true
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
