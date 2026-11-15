# Update

Update a custom CA certificate.

The agent deploys the PEM file again and runs `update-ca-certificates` to make
the trust store again.

```bash
osapi client node certificate update
```

## Examples

```bash
osapi client node certificate update --target web-01 \
  --name internal-ca --object internal-ca-v2
osapi client node certificate update --target _all \
  --name internal-ca --object internal-ca-v2
osapi client node certificate update --target web-01 \
  --name internal-ca --object internal-ca-v2 --json
```

## Result

The command fails if the certificate does not exist. If the content did not
change (the same SHA), the command returns `changed: false`.

## Flags

| Flag       | Description                                        | Default |
| ---------- | -------------------------------------------------- | ------- |
| `--name`   | Certificate name (required)                        |         |
| `--object` | Object store reference for the PEM file (required) |         |

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
