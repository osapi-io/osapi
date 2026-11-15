# Delete

Delete NTP configuration.

The agent removes the drop-in file from `/etc/chrony/conf.d/`, then reloads
chrony. The system then uses the other chrony configuration that stays.

```bash
osapi client node ntp delete
```

## Examples

```bash
osapi client node ntp delete --target web-01
osapi client node ntp delete --target _all
osapi client node ntp delete --target web-01 --json
```

## Result

The command fails if the configuration does not exist.

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
