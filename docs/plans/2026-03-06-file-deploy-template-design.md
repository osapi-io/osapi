# File Deploy & Template Rendering Design

## Context

OSAPI manages system configuration through async jobs. Current operations (DNS,
disk, memory, commands) send small JSON payloads through NATS KV. File
management — deploying config files, rendering templates with per-host facts —
requires transferring larger blobs and tracking deployed state for idempotency.

Ansible's approach transfers the full file every run to verify whether it
changed. We want SHA-based idempotency: compute the hash of what should be on
disk, compare against what was last deployed, and skip the transfer when nothing
changed.

## Goals

- Upload files to a central store (NATS Object Store) via the REST API
- Deploy files to agent hosts with mode, owner, and group control
- Render Go `text/template` files agent-side using live facts + user vars
- SHA-based idempotency — skip transfer when content hasn't changed
- Report `changed: true/false` so orchestrator guards (`OnlyIfChanged`) work
- **Shared primitive** — the Object Store layer is reusable by future providers
  (firmware, packages, certs, scripts), not tied to the file provider

## Design Decisions

- **Approach B: single operation with `content_type` flag.** One `file.deploy`
  operation. A `content_type` field (`raw` or `template`) controls whether the
  agent renders content before writing. SHA is computed on the **rendered
  output** for templates, so fact changes trigger redeployment.
- **NATS Object Store** for blob storage. It handles chunking automatically (KV
  has a ~1MB value limit). Files are uploaded once and pulled by agents on
  demand.
- **Dedicated `file-state` KV bucket** for SHA tracking. Keyed by
  `<hostname>.<sha256-of-path>`. No TTL — deployed state persists until
  explicitly removed. Separate from `agent-state` to keep concerns clean.
  Visible to the API server for fleet-wide deployment status.
- **Agent-side template rendering.** Raw Go template stored in Object Store.
  Agent renders locally using its cached facts + user-supplied vars. Consistent
  with how `@fact.*` resolution works today — each host gets its own output.
- **Mode + owner/group in job params.** Agent sets permissions after writing.
  Defaults to umask/current user when not specified.

## Architecture

### Shared Object Store Primitive

The Object Store client is a **shared agent dependency** — injected at startup
like `execManager`, `hostProvider`, or `factsKV`. Any provider can use it to
pull blobs.

```
┌─────────────────────────────────┐
│         Object Store            │  ← shared NATS resource
│    (file-objects bucket)        │
└──────────┬──────────────────────┘
           │
     ┌─────┴──────┐
     │ Agent      │
     │  .objStore │  ← injected handle
     └─────┬──────┘
           │
  ┌────────┼────────────┬───────────────┐
  │        │            │               │
file    firmware     package         cert
provider  provider    provider       provider
(now)    (future)    (future)       (future)
```

Future providers that would consume the Object Store:

| Provider          | Operation                                      | Usage |
| ----------------- | ---------------------------------------------- | ----- |
| `firmware.update` | Pull binary, run flash tool                    |
| `package.install` | Pull `.deb`/`.rpm`, install via `dpkg`/`rpm`   |
| `cert.deploy`     | Pull TLS cert/key, write with restricted perms |
| `script.run`      | Pull script file, execute with args            |

Each provider owns its domain logic but shares: Object Store download, SHA
comparison, and state tracking from the `file-state` KV bucket.

### Data Flow

**Upload phase** (new REST endpoint):

1. Client sends file content via `POST /file` with metadata (name)
2. API server stores content in NATS Object Store (`file-objects`)
3. Returns object reference: `{name, sha256, size}`

**Deploy phase** (job system — `file.deploy` operation):

1. Client creates job with `file.deploy` targeting host(s)
2. Job data: object name, destination path, mode, owner, group, content_type,
   optional template vars
3. Agent pulls object from Object Store
4. If `content_type: "template"` — renders with Go `text/template`
5. Computes SHA of final content (rendered or raw)
6. Checks `file-state` KV — if SHA matches, returns `changed: false`
7. If different — writes file, sets perms, updates state KV, returns
   `changed: true`

**Status check** (read-only — `file.status` operation):

1. Agent reads local file SHA, compares against `file-state` KV
2. Reports: in-sync, drifted, or missing

## Data Structures

### NATS Configuration

```yaml
nats:
  objects:
    bucket: 'file-objects'
    max_bytes: 524288000 # 500 MiB
    storage: 'file'
    replicas: 1

  file_state:
    bucket: 'file-state'
    storage: 'file'
    replicas: 1
    # No TTL — deployed file state persists
```

### File State KV Entry

Keyed by `<hostname>.<sha256-of-path>`:

```json
{
  "object_name": "nginx.conf",
  "path": "/etc/nginx/nginx.conf",
  "sha256": "abc123...",
  "mode": "0644",
  "owner": "root",
  "group": "root",
  "deployed_at": "2026-03-06T...",
  "content_type": "raw"
}
```

### Job Request Data (`file.deploy`)

```json
{
  "object_name": "nginx.conf",
  "path": "/etc/nginx/nginx.conf",
  "mode": "0644",
  "owner": "root",
  "group": "root",
  "content_type": "template",
  "vars": {
    "worker_count": 4,
    "upstream": "10.0.0.5"
  }
}
```

### Template Rendering Context

```go
type TemplateContext struct {
    Facts    *job.FactsRegistration
    Vars     map[string]any
    Hostname string
}
```

Example template:

```
worker_processes {{ .Vars.worker_count }};
# Running on {{ .Hostname }} ({{ .Facts.Architecture }})
server {{ .Vars.upstream }}:{{ if eq .Facts.Architecture "arm64" }}8081{{ else }}8080{{ end }};
```

## API Endpoints

| Method                | Path         | Permission                  | Description |
| --------------------- | ------------ | --------------------------- | ----------- |
| `POST /file`          | `file:write` | Upload file to Object Store |
| `GET /file`           | `file:read`  | List stored objects         |
| `GET /file/{name}`    | `file:read`  | Get object metadata         |
| `DELETE /file/{name}` | `file:write` | Remove stored object        |

Deploy and status go through the existing job system as `file.deploy` and
`file.status` operations. No new job endpoints needed.

### Permissions

New permissions: `file:read`, `file:write`. Added to `admin` and `write`
built-in roles.

## Agent-Side Architecture

The agent gets two new dependencies:

- **`objectStore`** — NATS Object Store handle. Any provider can use it.
- **`fileStateKV`** — dedicated KV for tracking deployed file SHAs.

The file provider implements:

- `Deploy(req) → (Result, error)` — pull from Object Store, optionally render
  template, SHA compare, write file, set perms, update state
- `Status(req) → (Result, error)` — read-only: compare local file SHA against
  state KV

The processor dispatch adds a `file` category alongside `node`, `network`,
`command`.

## SDK & Orchestrator Integration

### SDK (`osapi-sdk`)

New `FileService`:

- `Upload(ctx, name, content)` — upload to Object Store
- `List(ctx)` — list stored objects
- `Get(ctx, name)` — get object metadata
- `Delete(ctx, name)` — remove object

Deploy uses existing `Job.Create()` with operation `file.deploy`.

### Orchestrator (`osapi-orchestrator`)

```go
o := orchestrator.New(client)

upload := o.FileUpload("nginx.conf", "./local/nginx.conf.tmpl")
deploy := o.FileTemplate("_all", "nginx.conf", "/etc/nginx/nginx.conf",
    map[string]any{"worker_count": 4},
    orchestrator.WithMode("0644"),
    orchestrator.WithOwner("root", "root"),
).After(upload)

reload := o.CommandExec("_all", "nginx", []string{"-s", "reload"}).
    After(deploy).
    OnlyIfChanged()
```

- `FileDeploy()` — raw file deploy step
- `FileTemplate()` — deploy with `content_type: "template"`
- `OnlyIfChanged` works naturally via `changed` response field
- Template vars support `@fact.*` references (resolved agent-side)

## Verification

After implementation:

```bash
# Upload a file
osapi client file upload --name nginx.conf --file ./nginx.conf

# Deploy raw file
osapi client node file deploy \
  --object nginx.conf \
  --path /etc/nginx/nginx.conf \
  --mode 0644 --owner root --group root \
  --target _all

# Deploy template
osapi client node file deploy \
  --object nginx.conf.tmpl \
  --path /etc/nginx/nginx.conf \
  --content-type template \
  --var worker_count=4 \
  --mode 0644 --owner root --group root \
  --target _all

# Check status (idempotent re-run should show changed: false)
osapi client node file status \
  --path /etc/nginx/nginx.conf \
  --target _all
```
