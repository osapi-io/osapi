# Delete

Remove a custom CA certificate.

The agent removes the certificate file from `/usr/local/share/ca-certificates/`.

```bash
osapi client node certificate delete
```

## Examples

```bash
osapi client node certificate delete --target web-01 \
  --name internal-ca
osapi client node certificate delete --target _all \
  --name internal-ca
osapi client node certificate delete --target web-01 \
  --name internal-ca --json
```

## Flags

| Flag     | Description                           | Default |
| -------- | ------------------------------------- | ------- |
| `--name` | Certificate name to delete (required) |         |

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
