# Deploy

Deploy a file from the Object Store to the target host's filesystem. SHA-256
idempotency ensures unchanged files are not rewritten.

```bash
$ osapi client node file deploy --object app.conf --path /etc/app/app.conf

  Job ID:   550e8400-e29b-41d4-a716-446655440000
  Hostname: server1
  Changed:  true
```

Deploy with file permissions:

```bash
$ osapi client node file deploy \
    --object app.conf \
    --path /etc/app/app.conf \
    --mode 0644 \
    --owner root \
    --group root
```

Deploy a template with variables. Each agent renders the template with its own
facts and hostname:

```bash
$ osapi client node file deploy \
    --object app.conf.tmpl \
    --path /etc/app/app.conf \
    --content-type template \
    --var listen_address=0.0.0.0:8080 \
    --var max_workers=16 \
    --target _all
```

When targeting all hosts, the CLI prompts for confirmation:

```bash
$ osapi client node file deploy --object app.conf --path /etc/app/app.conf --target _all

  This will deploy the file to ALL hosts. Continue? [y/N] y

  Job ID:   550e8400-e29b-41d4-a716-446655440000
  Hostname: server1
  Changed:  true
```

Target by label to deploy to a group of servers:

```bash
$ osapi client node file deploy \
    --object nginx.conf \
    --path /etc/nginx/nginx.conf \
    --target group:web
```

See [File Management](../../../../../features/file-management.md) for details on
template rendering and SHA-based idempotency.

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node file deploy --object app.conf --path /etc/app/app.conf --json
```

## Flags

| Flag             | Description                                              | Default |
| ---------------- | -------------------------------------------------------- | ------- |
| `--object`       | Name of the file in the Object Store (**required**)      |         |
| `--path`         | Destination path on the target filesystem (**required**) |         |
| `--content-type` | Content type: `raw` or `template`                        | `raw`   |
| `--mode`         | File permission mode (e.g., `0644`)                      |         |
| `--owner`        | File owner user                                          |         |
| `--group`        | File owner group                                         |         |
| `--var`          | Template variable as `key=value` (repeatable)            | `[]`    |
| `-T, --target`   | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`     | Output raw JSON response                                 |         |
