# Update

Replace a file in the Object Store.

The command replaces the content of a stored file, or stores it when the name is
new. The same content changes nothing (`Changed: false`).

```bash
osapi client file update [flags]
```

## Examples

```bash
osapi client file update --name app.conf --file /tmp/app.conf
```

## Permission

It needs the `file:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag             | Description                                                                                                                                  | Default |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `--content-type` | How the file should be treated during deploy. raw writes bytes as-is; template renders with Go text/template and agent facts (raw, template) | `raw`   |
| `--file`         | Path to the local file with the file content (required)                                                                                      |         |
| `--name`         | The name of the file in the Object Store (required)                                                                                          |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
