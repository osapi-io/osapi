# Deploy

Deploy a file from Object Store to a host.

The agent compares SHA-256 checksums, so it does not write a file again when the
file did not change. \[File
management](../../../../../features/file-management.md) tells how the agent
renders a template.

```bash
osapi client node file deploy
```

## Examples

```bash
osapi client node file deploy --object app.conf --path /etc/app/app.conf
osapi client node file deploy \
  --object app.conf \
  --path /etc/app/app.conf \
  --mode 0644 \
  --owner root \
  --group root
osapi client node file deploy \
  --object app.conf.tmpl \
  --path /etc/app/app.conf \
  --content-type template \
  --var listen_address=0.0.0.0:8080 \
  --var max_workers=16 \
  --target _all
```

## Flags

| Flag             | Description                                          | Default |
| ---------------- | ---------------------------------------------------- | ------- |
| `--content-type` | Content type: raw or template (default raw)          | `raw`   |
| `--group`        | File owner group                                     |         |
| `--mode`         | File permission mode (e.g., 0644)                    |         |
| `--object`       | Name of the file in the Object Store (required)      |         |
| `--owner`        | File owner user                                      |         |
| `--path`         | Destination path on the target filesystem (required) |         |
| `--var`          | Template variable as key=value (repeatable)          |         |

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
