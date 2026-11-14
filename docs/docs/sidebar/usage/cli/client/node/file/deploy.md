# Deploy

Deploy a file from the Object Store to the filesystem of the target host. The
agent compares SHA-256 checksums, so it does not write a file again when the
file did not change.

```bash
$ osapi client node file deploy --object app.conf --path /etc/app/app.conf

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  server1   changed  true

  1 host: 1 changed
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
facts and its hostname:

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

When you set `--content-type template`, the agent processes the content as a Go
[text/template](https://pkg.go.dev/text/template) before it writes the file to
disk. The template context has three top-level fields:

| Field       | Type             | Description                                |
| ----------- | ---------------- | ------------------------------------------ |
| `.Facts`    | `map[string]any` | Agent's collected system facts             |
| `.Vars`     | `map[string]any` | User-supplied variables from `--var` flags |
| `.Hostname` | `string`         | Target agent's hostname                    |

### Available Facts

Each agent collects facts automatically. The facts are all the fields that the
agent registers: `architecture`, `kernel_version`, `cpu_count`, `fqdn`,
`service_mgr`, `package_mgr`, `primary_interface`, `interfaces`, `routes`, and
the custom facts. Get them with `index`:

```text
arch={{ index .Facts "architecture" }}
cpus={{ index .Facts "cpu_count" }}
fqdn={{ index .Facts "fqdn" }}
```

### Template Examples

A simple variable:

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

Configuration for each host, from facts:

```text
# Generated for {{ .Hostname }}
server_name = {{ .Hostname }}
arch = {{ index .Facts "architecture" }}
cpus = {{ index .Facts "cpu_count" }}
```

A loop over a list variable. The values of `--var` are strings, so give lists
through the SDK or the orchestrator:

```text
{{ range .Vars.servers }}
upstream {{ . }};
{{ end }}
```

Target a label to deploy to a group of servers:

```bash
$ osapi client node file deploy \
    --object nginx.conf \
    --path /etc/nginx/nginx.conf \
    --target group:web
```

Refer to [File Management](../../../../../features/file-management.md). It tells
how the agent renders templates, and how the SHA checksums keep deploys
idempotent.

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
