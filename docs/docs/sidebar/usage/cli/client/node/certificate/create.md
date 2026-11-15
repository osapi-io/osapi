# Create

Add a custom CA certificate.

First, upload the PEM content to the Object Store. The agent writes the
certificate to `/usr/local/share/ca-certificates/{name}.crt`. Then
`update-ca-certificates` makes the system trust store again.

```bash
osapi client node certificate create
```

## Examples

```bash
osapi client node certificate create --target web-01 \
  --name internal-ca --object internal-ca
osapi client node certificate create --target _all \
  --name internal-ca --object internal-ca
osapi client node certificate create --target web-01 \
  --name internal-ca --object internal-ca --json
```

## Result

If the certificate file has the same content, the command returns
`changed: false`. If the file has different content, the agent writes it again.

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
