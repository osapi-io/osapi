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

1. The CLI (or SDK) computes a SHA-256 of the local file and queries the Object
   Store to check whether it already holds the same content. If the SHA matches,
   the upload is skipped entirely (no bytes sent over the network).
2. If the file is new or the SHA differs, the CLI sends the file via a multipart
   upload.
3. On the server side, if a file with the same name already exists and the
   content differs, the server rejects the upload with **409 Conflict** unless
   `?force=true` is passed.
4. If the content is identical, the server returns `changed: false` without
   rewriting the object.
5. With `--force`, both the SDK pre-check and the server-side digest guard are
   bypassed — the file is always written and `changed: true` is returned.

### File Deploy Flow

Deploy follows the standard [job processing flow](job-system.md). The agent
fetches the file from Object Store, renders it if it is a template, and compares
the result with **the file on disk**. If they differ — because the file is
absent, or because somebody edited it — the agent writes it and records the new
SHA-256.

The comparison is against the disk, not against what the last deploy recorded. A
config edited by hand still has the SHA the record remembers, so comparing
against the record would call that unchanged and walk away from the drift.

The write is atomic. Content goes to a temporary file beside the target, is
given the requested mode, and is renamed over it, so a reader — an `sshd`, an
`nginx -t`, a `systemctl daemon-reload` — sees either the old file or the new
one, never half of either. A write that fails part way leaves the target
untouched and removes the temporary file.

**Mode, owner and group are applied even when the content has not changed**, so
a deploy that changes only the permissions takes effect and reports
`changed: true`. Each is compared against the file itself:

- The mode is compared with the permission bits on disk, so a mode changed by
  anything is corrected.
- The owner and group are compared with the file's actual uid and gid, after
  resolving the requested names on the host. A name the host does not know fails
  the deploy rather than passing silently; a numeric id is accepted as itself,
  which is what a container image without a `passwd` entry needs. When they
  differ, the agent runs `chown` through its privilege escalation.
- An absent `mode` means "leave the permissions alone" rather than 0644. The
  default applies to a file being created, because applying it to a file already
  on disk would quietly widen permissions someone else set. The same holds for
  an absent `owner` or `group`.

What osapi records is a log of what it did, never evidence of what is there now.
Every decision above reads the system.

You can target a specific host, broadcast to all hosts with `_all`, or route by
label.

### File Undeploy Flow

```bash
osapi client node file undeploy --target HOST --path /etc/app/app.conf
```

Undeploy follows the standard [job processing flow](job-system.md). The agent
removes the file from the filesystem but preserves the file-state KV record so
the operation is auditable. A subsequent deploy will write the file even if the
content has not changed (since the file is now absent). If the file does not
exist on disk, the operation returns `changed: false`.

### SHA-Based Idempotency

Every deploy computes a SHA-256 of the content it was asked to deploy and
compares it with a SHA-256 of the file on disk. If they match, the file is not
rewritten. This makes repeated deploys safe and efficient -- only actual
differences hit the filesystem, and a file somebody edited counts as a
difference. The permissions are checked either way, as described under File
Deploy Flow above: identical content does not mean an identical file.

The file-state KV records what was deployed, for `status`, staleness detection
and audit. It is not consulted to decide whether to write.

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

| Key                 | Type     | Description                  | Example          |
| ------------------- | -------- | ---------------------------- | ---------------- |
| `architecture`      | string   | CPU architecture             | `amd64`, `arm64` |
| `kernel_version`    | string   | OS kernel version            | `6.8.0-51`       |
| `cpu_count`         | number   | Logical CPU count            | `8`              |
| `fqdn`              | string   | Fully qualified domain name  | `web-01.lan`     |
| `service_mgr`       | string   | Init system                  | `systemd`        |
| `package_mgr`       | string   | System package manager       | `apt`            |
| `containerized`     | boolean  | Running inside a container   | `true`           |
| `primary_interface` | string   | Default route interface name | `eth0`           |
| `interfaces`        | []object | Network interfaces           | _(see below)_    |
| `routes`            | []object | IP routing table             | _(see below)_    |

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
the deploy to **fail with an error** rather than silently rendering
`<no value>`.

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

## Staleness Detection

When an object is re-uploaded to the Object Store with new content, existing
deployments become stale — the deployed file no longer matches the source. The
`file stale` command detects this by comparing the SHA-256 hash of each
deployment against the current object content.

```bash
osapi client file stale
```

This is a controller-side check that does not contact agents. It compares the
file-state KV (which tracks what was deployed) against the Object Store (which
has the current content). Use `file deploy` to bring stale deployments up to
date.

Stale detection covers all providers that use the file provider:

- Service unit files
- CA certificates
- Cron scripts
- Direct file deployments

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

File operations require `file:*` permissions. The `admin` and `write` roles
include both `file:read` and `file:write`. The `read` role includes only
`file:read`. See the [API reference](/category/api) for the permission required
by each endpoint.

## Related

- [System Facts](system-facts.md) -- facts available in template context
- [Job System](job-system.md) -- how async job processing works
- [Authentication & RBAC](authentication.md) -- permissions and roles
- [Architecture](../architecture/architecture.md) -- system design overview
