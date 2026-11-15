# Create

Create NTP configuration.

The agent writes the server list to `/etc/chrony/conf.d/osapi-ntp.conf` and
reloads chrony immediately.

```bash
osapi client node ntp create
```

## Examples

```bash
osapi client node ntp create --target web-01 \
  --servers 0.pool.ntp.org --servers 1.pool.ntp.org
osapi client node ntp create --target _all \
  --servers 0.pool.ntp.org --servers 1.pool.ntp.org
osapi client node ntp create --target _all \
  --servers 0.pool.ntp.org
```

## Result

If osapi already manages it, the command returns `changed: false`. To change the
server list, use `update`.

## Flags

| Flag        | Description                     | Default |
| ----------- | ------------------------------- | ------- |
| `--servers` | NTP server addresses (required) |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
