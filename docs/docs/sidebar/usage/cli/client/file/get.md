# Get

Get file metadata.

Get metadata for a specific file in the Object Store.

```bash
osapi client file get [flags]
```

## Examples

```bash
osapi client file get --name app.conf
osapi client file get --name app.conf --json
```

## Permission

It needs the `file:read` permission. `osapi token roles list` lists the roles
that grant it.

## Flags

| Flag     | Description                                         | Default |
| -------- | --------------------------------------------------- | ------- |
| `--name` | The name of the file in the Object Store (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
