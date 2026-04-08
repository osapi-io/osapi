# List

List all network interfaces on a target host:

```bash
$ osapi client node network interface list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME    IPV4                DHCP    PRIMARY
  eth0    192.168.1.100/24    static  yes
  lo      127.0.0.1/8
```

Target all hosts to list interfaces across the fleet:

```bash
$ osapi client node network interface list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  NAME    IPV4                DHCP    PRIMARY
  eth0    192.168.1.100/24    static  yes

  web-02
  NAME    IPV4                DHCP    PRIMARY
  eth0    192.168.1.200/24    static  yes
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network interface list --target web-01 --json
{"results":[{"hostname":"web-01","status":"ok","interfaces":[
{"name":"eth0","dhcp4":false,"dhcp6":false,
"addresses":["192.168.1.100/24"],"gateway4":"192.168.1.1",
"mac_address":"52:54:00:ab:cd:ef","state":"up"}
]}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
