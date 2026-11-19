# Upload

Upload a file to the Object Store.

An upload of the same content changes nothing (`Changed: false`). If the file
exists with different content, the command rejects the upload with 409 Conflict.
`file update` replaces the file.

```bash
osapi client file upload [flags]
```

## Examples

```bash
osapi client file upload --name app.conf --file /tmp/app.conf
osapi client file upload --name app.conf.tmpl --file /tmp/app.conf.tmpl --content-type template
```

## Permission

It needs the `file:write` permission. `osapi token roles list` lists the roles
that grant it.

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
