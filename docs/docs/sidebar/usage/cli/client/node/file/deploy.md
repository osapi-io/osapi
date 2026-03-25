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

## Template Rendering

When `--content-type template` is set, file content is processed as a Go
[text/template](https://pkg.go.dev/text/template) before being written to disk.
The template context provides three top-level fields:

| Field       | Type             | Description                                |
| ----------- | ---------------- | ------------------------------------------ |
| `.Facts`    | `map[string]any` | Agent's collected system facts             |
| `.Vars`     | `map[string]any` | User-supplied variables from `--var` flags |
| `.Hostname` | `string`         | Target agent's hostname                    |

### Available Facts

Facts are collected automatically by each agent and include all fields from the
agent's fact registration: `architecture`, `kernel_version`, `cpu_count`,
`fqdn`, `service_mgr`, `package_mgr`, `primary_interface`, `interfaces`,
`routes`, plus any custom facts. Access them with `index`:

```text
arch={{ index .Facts "architecture" }}
cpus={{ index .Facts "cpu_count" }}
fqdn={{ index .Facts "fqdn" }}
```

### Template Examples

Simple variable substitution:

```text
listen = {{ .Vars.listen_address }}
workers = {{ .Vars.max_workers }}
```

Conditionals:

```text
{{ if eq .Vars.env "prod" }}
log_level = warn
{{ else }}
log_level = debug
{{ end }}
```

Host-specific configuration using facts:

```text
# Generated for {{ .Hostname }}
server_name = {{ .Hostname }}
arch = {{ index .Facts "architecture" }}
cpus = {{ index .Facts "cpu_count" }}
```

Iterating over a list variable (`--var` values are strings, so pass lists via
the SDK or orchestrator):

```text
{{ range .Vars.servers }}
upstream {{ . }};
{{ end }}
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
| `-T, --target`   | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`     | Output raw JSON response                                 |         |
