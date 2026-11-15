# Get

Get file metadata.

```bash
osapi client file get
```

## Examples

```bash
osapi client file get --name app.conf
osapi client file get --name app.conf --json
```

## Flags

| Flag     | Description                                     | Default |
| -------- | ----------------------------------------------- | ------- |
| `--name` | Name of the file in the Object Store (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
