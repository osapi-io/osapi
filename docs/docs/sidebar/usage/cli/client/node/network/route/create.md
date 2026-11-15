# Create

Create routes for an interface.

Give each route as `TO:VIA` or `TO:VIA:METRIC`. `TO` is the destination in CIDR
notation. `VIA` is the IP address of the gateway. `METRIC` is the priority of
the route, and a lower value has priority.

```bash
osapi client node network route create
```

## Examples

```bash
osapi client node network route create \
  --target web-01 --interface eth0 \
  --route 10.0.0.0/8:192.168.1.1
osapi client node network route create \
  --target web-01 --interface eth0 \
  --route 10.0.0.0/8:192.168.1.1 \
  --route 172.16.0.0/12:192.168.1.1:100
osapi client node network route create \
  --target _all --interface eth0 \
  --route 10.0.0.0/8:192.168.1.1
```

## Result

If osapi already manages a route configuration for that interface, the command
returns `changed: false` and does not change the file. To change it, use
`update`. The command fails if a route is a default route.

## Flags

| Flag          | Description                                                    | Default |
| ------------- | -------------------------------------------------------------- | ------- |
| `--interface` | Interface name (required)                                      |         |
| `--route`     | Route in TO:VIA or TO:VIA:METRIC format (repeatable, required) |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
