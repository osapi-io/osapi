# Get

Get configuration details for a specific network interface:

```bash
$ osapi client node network interface get \
    --target web-01 --name eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  Hostname     web-01
  Name         eth0
  DHCP4        false
  DHCP6        false
  Addresses    192.168.1.100/24
  Gateway4     192.168.1.1
  Gateway6
  MTU          1500
  MAC Address  52:54:00:ab:cd:ef
  Wake-on-LAN  false
  Managed      true
  State        up
```

When targeting all hosts:

```bash
$ osapi client node network interface get \
    --target _all --name eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  Hostname     web-01
  Name         eth0
  ...

  Hostname     web-02
  Name         eth0
  ...
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network interface get \
    --target web-01 --name eth0 --json
{"results":[{"hostname":"web-01","status":"ok","interface":
{"name":"eth0","dhcp4":false,"dhcp6":false,
"addresses":["192.168.1.100/24"],"gateway4":"192.168.1.1",
"mtu":1500,"mac_address":"52:54:00:ab:cd:ef",
"managed":true,"state":"up"}}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Interface name                                           | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
