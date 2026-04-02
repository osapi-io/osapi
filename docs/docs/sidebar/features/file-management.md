---
sidebar_position: 9
---

# File Management

OSAPI can upload files to a central Object Store and deploy them to managed
hosts with SHA-based idempotency. File operations run through the
[job system](job-system.md), so the API server never writes to the filesystem
directly -- agents handle all deployment.

## What It Does

| Operation | Description                                            |
| --------- | ------------------------------------------------------ |
| Upload    | Store a file (base64-encoded) in the NATS Object Store |
| List      | List all files stored in the Object Store              |
| Get       | Retrieve metadata for a specific stored file           |
| Delete    | Remove a file from the Object Store                    |
| Deploy    | Deploy a file from Object Store to agent filesystem    |
| Undeploy  | Remove a deployed file from disk (state preserved)     |
| Status    | Check whether a deployed file is in-sync or drifted    |

**Upload / List / Get / Delete** manage files in the central NATS Object Store.
Files are stored by name and tracked with SHA-256 checksums. These operations
are synchronous REST calls -- they do not go through the job system.

**Deploy** creates an asynchronous job that fetches the file from the Object
Store and writes it to the target path on the agent's filesystem. Deploy
supports optional file permissions (mode, owner, group) and Go template
rendering.

**Undeploy** creates an asynchronous job that removes a previously deployed file
from the agent's filesystem. The file-state KV record is preserved so the
undeploy is auditable and a subsequent deploy can detect the change.

**Status** creates an asynchronous job that compares the current file on disk
against its expected SHA-256 from the file-state KV bucket. It reports one of
three states: `in-sync`, `drifted`, or `missing`.

## How It Works

### File Upload Flow

1. The CLI (or SDK) computes a SHA-256 of the local file and calls
   `GET /file/{name}` to check whether the Object Store already holds the same
   content. If the SHA matches, the upload is skipped entirely (no bytes sent
   over the network).
2. If the file is new (404) or the SHA differs, the CLI sends the file via
   multipart `POST /file`.
3. On the server side, if a file with the same name already exists and the
   content differs, the server rejects the upload with **409 Conflict** unless
   `?force=true` is passed.
4. If the content is identical, the server returns `changed: false` without
   rewriting the object.
5. With `--force`, both the SDK pre-check and the server-side digest guard are
   bypassed — the file is always written and `changed: true` is returned.

### File Deploy Flow

1. The CLI posts a deploy request specifying the Object Store file name, target
   path, and optional permissions.
2. The API server creates a job and publishes it to NATS.
3. An agent picks up the job, fetches the file from Object Store, and computes
   its SHA-256.
4. The agent checks the file-state KV for a previous deploy. If the SHA matches,
   the file is skipped (idempotent no-op).
5. If the content differs, the agent writes the file to disk and updates the
   file-state KV with the new SHA-256.
6. The result (changed, SHA-256, path) is written back to NATS KV.

You can target a specific host, broadcast to all hosts with `_all`, or route by
label.

### File Undeploy Flow

```bash
osapi client node file undeploy --target HOST --path /etc/app/app.conf
```

1. The CLI posts an undeploy request specifying the target path.
2. The API server creates a job and publishes it to NATS.
3. An agent picks up the job and removes the file from the filesystem.
4. The file-state KV record is preserved -- it is not deleted. This keeps the
   undeploy auditable and ensures a subsequent deploy will write the file even
   if the content has not changed (since the file is now absent).
5. The result (changed, path) is written back to NATS KV.

If the file does not exist on disk when undeploy runs, the operation returns
`changed: false`.

### SHA-Based Idempotency

Every deploy operation computes a SHA-256 of the file content and compares it
against the previously deployed SHA stored in the file-state KV bucket. If the
hashes match, the file is not rewritten. This makes repeated deploys safe and
efficient -- only actual changes hit the filesystem.

The file-state KV has no TTL, so deploy state persists indefinitely until
explicitly removed.

## Protected Objects

Files stored under the `osapi/` name prefix are protected. Both uploads and
deletes to `osapi/*` names return **403 Forbidden**. These objects are managed
exclusively by osapi itself — the agent seeds them on startup from embedded
templates and updates them automatically when a new osapi version ships with
changes.

Protected objects are used by meta providers such as the cron provider, which
references them at deploy time. The `osapi/` prefix is reserved; use any other
prefix for your own files.

## Template Rendering

When `content_type` is set to `template`, the file content is processed as a Go
`text/template` before being written to disk. The template context provides
three top-level fields:

| Field       | Description                            |
| ----------- | -------------------------------------- |
| `.Facts`    | Agent's collected system facts (map)   |
| `.Vars`     | User-supplied template variables (map) |
| `.Hostname` | Target agent's hostname (string)       |

### Example Template

A configuration file that adapts to each host:

```text
# Generated for {{ .Hostname }}
listen_address = {{ .Vars.listen_address }}
workers = {{ .Facts.cpu_count }}
arch = {{ .Facts.architecture }}
```

Deploy it with template variables:

```bash
osapi client node file deploy \
    --object-name app.conf.tmpl \
    --path /etc/app/app.conf \
    --content-type template \
    --var listen_address=0.0.0.0:8080 \
    --target _all
```

Each agent renders the template with its own facts and hostname, so the same
template produces host-specific configuration across a fleet.

### Available Fact Keys

Facts are exposed as a flat map via JSON round-tripping of the agent's
`FactsRegistration`. Use dot-syntax (`.Facts.key`) for keys that are valid Go
identifiers:

| Key                  | Type     | Description                    | Example          |
| -------------------- | -------- | ------------------------------ | ---------------- |
| `architecture`       | string   | CPU architecture               | `amd64`, `arm64` |
| `kernel_version`     | string   | OS kernel version              | `6.8.0-51`       |
| `cpu_count`          | number   | Logical CPU count              | `8`              |
| `fqdn`              | string   | Fully qualified domain name    | `web-01.lan`     |
| `service_mgr`       | string   | Init system                    | `systemd`        |
| `package_mgr`       | string   | System package manager         | `apt`            |
| `containerized`     | boolean  | Running inside a container     | `true`           |
| `primary_interface` | string   | Default route interface name   | `eth0`           |
| `interfaces`        | []object | Network interfaces             | _(see below)_    |
| `routes`            | []object | IP routing table               | _(see below)_    |

Access scalar facts with dot-syntax:

```text
arch = {{ .Facts.architecture }}
cpus = {{ .Facts.cpu_count }}
```

For keys with underscores, both `{{ .Facts.kernel_version }}` and
`{{ index .Facts "kernel_version" }}` work.

:::caution Missing key behavior

Templates use Go's `missingkey=error` option. Accessing a key that doesn't exist
via **dot-syntax** (e.g., `{{ .Vars.missing }}` or `{{ .Facts.bogus }}`) causes
the deploy to **fail with an error** rather than silently rendering `<no value>`.

However, `{{ index .Facts "nonexistent" }}` uses Go's built-in `index` function,
which returns the zero value for the map's value type — rendering `<no value>`
without an error. **Prefer dot-syntax over `index`** for fact access so that
typos are caught at deploy time.

:::

### Meta Provider Templates

Domains that use file deployment (service management, certificate management)
inherit template support automatically. When the uploaded object has
`content_type: template`, the file provider renders it at deploy time — the meta
provider does not need to specify the content type.

For example, a systemd unit file template:

```text
[Unit]
Description=App on {{ .Hostname }}

[Service]
ExecStart=/usr/bin/app --cpus {{ .Facts.cpu_count }}
```

Upload as a template, then deploy via the service API:

```bash
osapi client file upload \
    --name my-unit --file app.service --content-type template

osapi client node service create \
    --target web-01 --name my-app --object my-unit
```

## Configuration

File management uses two NATS infrastructure components in addition to the
general job infrastructure:

- **Object Store** (`nats.objects`) -- stores uploaded file content. Configured
  with bucket name, max size, storage backend, and chunk size.
- **File State KV** (`nats.file_state`) -- tracks deploy state (SHA-256, path,
  timestamps) per host. Has no TTL -- state persists until explicitly removed.

See [Configuration](../usage/configuration.md) for the full reference.

```yaml
nats:
  objects:
    bucket: 'file-objects'
    max_bytes: 104857600
    storage: 'file'
    replicas: 1
    max_chunk_size: 262144

  file_state:
    bucket: 'file-state'
    storage: 'file'
    replicas: 1
```

## Permissions

| Endpoint                              | Permission   |
| ------------------------------------- | ------------ |
| `POST /file` (upload)                 | `file:write` |
| `GET /file` (list)                    | `file:read`  |
| `GET /file/{name}` (get)              | `file:read`  |
| `DELETE /file/{name}` (delete)        | `file:write` |
| `POST /node/{hostname}/file/deploy`   | `file:write` |
| `POST /node/{hostname}/file/undeploy` | `file:write` |
| `POST /node/{hostname}/file/status`   | `file:read`  |

The `admin` and `write` roles include both `file:read` and `file:write`. The
`read` role includes only `file:read`.

## Related

- [System Facts](system-facts.md) -- facts available in template context
- [Job System](job-system.md) -- how async job processing works
- [Authentication & RBAC](authentication.md) -- permissions and roles
- [Architecture](../architecture/architecture.md) -- system design overview
