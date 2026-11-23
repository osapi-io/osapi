---
sidebar_position: 9
---

# File management

osapi uploads files to a central Object Store and deploys them to managed hosts.
It compares SHA-256 checksums, so a deploy does not write a file that did not
change. File operations use the [job system](job-system.md). Thus the API server
does not write to a filesystem. The agents do all the deploys.

## What it does

| Operation | Description                                                 |
| --------- | ----------------------------------------------------------- |
| Upload    | Store a file (base64-encoded) in the NATS Object Store      |
| List      | List all files in the Object Store                          |
| Get       | Get the metadata of one stored file                         |
| Delete    | Remove a file from the Object Store                         |
| Deploy    | Deploy a file from the Object Store to the agent filesystem |
| Undeploy  | Remove a deployed file from disk. The state stays           |
| Status    | Check if a deployed file is in sync or drifted              |

**Upload, List, Get and Delete** manage the files in the central NATS Object
Store. osapi stores each file by name, with a SHA-256 checksum. These operations
are synchronous REST calls. They do not use the job system.

**Deploy** makes an asynchronous job. The job gets the file from the Object
Store and writes it to the target path on the filesystem of the agent. A deploy
can set the file permissions (mode, owner, group) and can render a Go template.
Deploy, undeploy and status are the update, delete and status methods of the
[file behavior](../reference/behaviors/file/file.md):
`osapi client node file update`, `osapi client node file delete` and
`osapi client node file status`. They run on each Linux host and on macOS.

**Undeploy** makes an asynchronous job that removes a deployed file from the
filesystem of the agent. The file-state KV record stays. Thus you can audit the
undeploy, and a later deploy can find the change.

**Status** makes an asynchronous job. The job compares the file on disk with its
expected SHA-256 from the file-state KV bucket. It reports one of four states:
`in-sync`, `drifted`, `missing` or `unmanaged`. `missing` is a file that osapi
deployed and that is not on the disk. `unmanaged` is a path that osapi did not
deploy, or undeployed. Status refuses a path that is a directory with a 400.

## How it works

### File upload flow

1. The CLI or the SDK calculates the SHA-256 of the local file. Then it asks the
   Object Store if it already has the same content. If the SHA is the same, the
   upload stops, and no bytes go over the network.
1. If the file is new or the SHA is different, the CLI sends the file in a
   multipart upload.
1. On the server, if a file with the same name exists and the content is
   different, the server rejects the upload with **409 Conflict**.
1. If the content is the same, the server returns `changed: false` and does not
   write the object again.
1. To replace a file, use `osapi client file update`, which sends
   `PUT /api/file/{name}`. The server writes the new content and answers 200, or
   201 when the name is new. The same content changes nothing.

### File deploy flow

A deploy follows the standard [job processing flow](job-system.md). The agent
gets the file from the Object Store and renders it if it is a template. Then it
compares the result with **the file on disk**. If they are different, because
the file is not there or because somebody changed it, the agent writes the file
and records the new SHA-256.

The agent compares with the disk, not with the record of the last deploy. A
configuration that somebody changed by hand still has the SHA that the record
keeps. A comparison with the record shows no change, and the drift stays.

The write is atomic. The agent writes the content to a temporary file next to
the target, gives it the requested mode, and renames it over the target. Thus a
reader, for example `sshd`, `nginx -t` or `systemctl daemon-reload`, sees the
old file or the new file, never a part of a file. If a write fails, the target
does not change, and the agent removes the temporary file.

**The agent applies the mode, the owner and the group also when the content did
not change.** Thus a deploy that changes only the permissions has an effect. It
reports `changed: true`. The agent compares each one with the file:

- It compares the mode with the permission bits on disk. Thus it corrects a mode
  that something else changed.
- It compares the owner and the group with the names that `stat` reads from the
  file. The owner and the group are names, not numeric ids. A name that the host
  does not know fails the deploy. When they are different, the agent runs
  `chown` through its privilege escalation.
- A `mode` must have three or four octal digits. A mode that the agent cannot
  parse fails the request. Before, such a mode became 0644, which gave more
  access than requested and still reported success.
- When the request has no `mode`, the agent does not change the permissions. It
  does not use 0644. The default applies only to a new file. On a file that
  exists, it can give more access than somebody else set. The same applies when
  the request has no `owner` or `group`.

The records of osapi are a log of what it did, not a proof of what is there now.
Each decision above reads the system.

You can target one host, broadcast to all hosts with `_all`, or target a label.

### File undeploy flow

```bash
osapi client node file delete --target HOST --path /etc/app/app.conf
```

An undeploy follows the standard [job processing flow](job-system.md). The agent
removes the file from the filesystem, but keeps the file-state KV record so that
you can audit the operation. A later deploy writes the file also when the
content did not change, because the file is not there now. If the file is not on
disk, the operation returns `changed: false`.

### SHA-based idempotency

Each deploy calculates the SHA-256 of the content to deploy, and compares it
with the SHA-256 of the file on disk. If they are the same, the agent does not
write the file again. Thus repeated deploys are safe and fast. Only real
differences go to the filesystem, and a file that somebody changed is a
difference. The agent checks the permissions in all cases, as File deploy flow
tells. The same content does not mean the same file.

The file-state KV records what osapi deployed, for `status`, for stale detection
and for audit. The agent does not use it to decide if it must write.

The file-state KV has no TTL. Thus the deploy state stays until you remove it.

## Protected objects

Files with the `osapi/` name prefix are protected. An upload or a delete of an
`osapi/*` name returns **403 Forbidden**. Only osapi manages these objects. The
agent adds them from embedded templates when it starts. It updates them
automatically when a new version of osapi changes them.

Meta providers use protected objects, for example the cron provider, which
refers to them at deploy time. The `osapi/` prefix is reserved. Use a different
prefix for your files.

## Template rendering

When `content_type` is `template`, the agent processes the content as a Go
`text/template` before it writes the file to disk. The template context has
three top-level fields:

| Field       | Description                                     |
| ----------- | ----------------------------------------------- |
| `.Facts`    | The system facts that the agent collected (map) |
| `.Vars`     | The template variables from the user (map)      |
| `.Hostname` | The hostname of the target agent (string)       |

### Example template

A configuration file that changes for each host:

```text
# Generated for {{ .Hostname }}
listen_address = {{ .Vars.listen_address }}
workers = {{ .Facts.cpu_count }}
arch = {{ .Facts.architecture }}
```

Deploy it with template variables:

```bash
osapi client node file update \
    --object-name app.conf.tmpl \
    --path /etc/app/app.conf \
    --content-type template \
    --vars listen_address=0.0.0.0:8080 \
    --target _all
```

Each agent renders the template with its own facts and hostname. Thus the same
template gives a different configuration for each host in a fleet.

### Conditions and loops

A template can use each action of Go `text/template`. A condition:

```text
{{ if eq .Vars.env "prod" }}
log_level = warn
{{ else }}
log_level = debug
{{ end }}
```

Each value of `.Vars` is a string, from the CLI, the SDK and the API. A key that
the template reads and the call does not give fails the deploy. A loop goes over
a list that a fact gives:

```text
{{ range .Facts.interfaces }}
# {{ .name }}
{{ end }}
```

### Available fact keys

The facts are a flat map. osapi makes it from the `FactsRegistration` of the
agent, through JSON. Use dot syntax (`.Facts.key`) for keys that are valid Go
identifiers:

| Key                 | Type     | Description                         | Example          |
| ------------------- | -------- | ----------------------------------- | ---------------- |
| `architecture`      | string   | CPU architecture                    | `amd64`, `arm64` |
| `kernel_version`    | string   | OS kernel version                   | `6.8.0-51`       |
| `cpu_count`         | number   | Logical CPU count                   | `8`              |
| `fqdn`              | string   | Fully qualified domain name         | `web-01.lan`     |
| `service_mgr`       | string   | Init system                         | `systemd`        |
| `package_mgr`       | string   | System package manager              | `apt`            |
| `containerized`     | boolean  | The agent runs in a container       | `true`           |
| `primary_interface` | string   | Name of the default route interface | `eth0`           |
| `interfaces`        | []object | Network interfaces                  | _(see below)_    |
| `routes`            | []object | IP routing table                    | _(see below)_    |

Get a scalar fact with dot syntax:

```text
arch = {{ .Facts.architecture }}
cpus = {{ .Facts.cpu_count }}
```

For keys with underscores, `{{ .Facts.kernel_version }}` and
`{{ index .Facts "kernel_version" }}` both work.

:::caution Missing key behavior

Templates use the Go `missingkey=error` option. If you use dot syntax for a key
that does not exist, for example `{{ .Vars.missing }}` or `{{ .Facts.bogus }}`,
the deploy **fails with an error**. It does not write `<no value>`.

But `{{ index .Facts "nonexistent" }}` uses the built-in Go `index` function. It
returns the zero value of the map value type, and writes `<no value>` without an
error. **Use dot syntax, not `index`**, to get facts. Then a typo fails at
deploy time.

:::

### Meta provider templates

The domains that deploy files, for example service management and certificate
management, get template support automatically. When the uploaded object has
`content_type: template`, the file provider renders it at deploy time. The meta
provider does not need to set the content type.

For example, a systemd unit file template:

```text
[Unit]
Description=App on {{ .Hostname }}

[Service]
ExecStart=/usr/bin/app --cpus {{ .Facts.cpu_count }}
```

Upload it as a template. Then deploy it through the service API:

```bash
osapi client file upload \
    --name my-unit --file app.service --content-type template

osapi client node service create \
    --target web-01 --name my-app --object my-unit
```

## Staleness detection

When you upload an object again with new content, the deployments of it become
stale. The deployed file is then different from the source. `file stale list`
finds this. It compares the SHA-256 of each deployment with the current content
of the object.

```bash
osapi client file stale list
```

The controller does this check, and it does not contact the agents. It compares
the file-state KV, which records what osapi deployed, with the Object Store,
which has the current content. Use `osapi client node file update` to update
stale deployments.

Stale detection includes all providers that use the file provider:

- Service unit files
- CA certificates
- Cron scripts
- Direct file deployments

## Configuration

File management uses two NATS parts in addition to the general job system:

- **Object Store** (`nats.objects`) keeps the uploaded file content. Configure
  its bucket name, maximum size, storage backend and chunk size.
- **File State KV** (`nats.file_state`) records the deploy state (SHA-256, path,
  times) for each host. It has no TTL. Thus the state stays until you remove it.

Refer to [Configuration](../usage/configuration.md) for all the settings.

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

The Permission column of [Methods](#methods) gives the permission of each
method.

## Related

- [System Facts](system-facts.md): the facts in the template context
- [Job System](job-system.md): how asynchronous jobs work
- [Authentication & RBAC](authentication.md): permissions and roles
- [Architecture](../architecture/architecture.md): an overview of the system
  design
