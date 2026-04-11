# Create

Create a new Netplan interface configuration on a target host. Fails if an
OSAPI-managed configuration for that interface already exists -- use `update` to
replace it:

```bash
$ osapi client node network interface create \
    --target web-01 --name eth0 \
    --address 192.168.1.100/24 \
    --gateway4 192.168.1.1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED
  web-01    changed  eth0    true

  1 host: 1 changed
```

Create an interface with DHCP:

```bash
$ osapi client node network interface create \
    --target web-01 --name eth1 --dhcp4

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED
  web-01    changed  eth1    true

  1 host: 1 changed
```

Broadcast to all hosts at once:

```bash
$ osapi client node network interface create \
    --target _all --name eth1 --dhcp4

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED
  web-01    changed  eth1    true
  web-02    changed  eth1    true

  2 hosts: 2 changed
```

When some hosts are skipped (e.g., macOS agents):

```bash
$ osapi client node network interface create \
    --target _all --name eth1 --dhcp4

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED
  web-01    changed  eth1    true
  mac-01    skip

  2 hosts: 1 changed, 1 skipped

  Details:
  mac-01    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network interface create \
    --target web-01 --name eth0 \
    --address 192.168.1.100/24 \
    --gateway4 192.168.1.1 --json
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
