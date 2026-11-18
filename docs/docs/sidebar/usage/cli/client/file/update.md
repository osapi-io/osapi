# Update

Replace a file in the Object Store.

The command replaces the content of a stored file, or stores it when the name is
new. The same content changes nothing (`Changed: false`).

```bash
osapi client file update
```

## Examples

```bash
osapi client file update --name app.conf --file /tmp/app.conf
```

## Permission

It needs the `file:write` permission. `osapi token roles` lists the roles that
grant it.

## Flags

| Flag             | Description                                            | Default |
| ---------------- | ------------------------------------------------------ | ------- |
| `--content-type` | File type: raw or template                             | `raw`   |
| `--file`         | Path to the local file with the new content (required) |         |
| `--name`         | Name of the file in the Object Store (required)        |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
