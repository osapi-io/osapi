# File-Backed Meta Providers Design

## Goal

Refactor the cron provider (and establish the pattern for future providers like
systemd, sysctl, apt sources) to delegate file writes to the file provider
instead of using raw `afero.WriteFile`. This gives all file-writing providers
SHA tracking, idempotency, drift detection, and template rendering for free.

Additionally, add `Undeploy` to the file provider (remove file from disk while
keeping the object in the store), and add `protected` object support so
system-managed templates cannot be deleted by users.

## Architecture

### Meta Provider Pattern

A meta provider is a domain-specific provider that writes files to well-known
paths. It does not write to the filesystem directly. Instead, it:

1. Determines the destination path and permissions based on domain rules
2. Delegates to the file provider's `Deploy()` method
3. Gets SHA tracking, idempotency, and template rendering for free

```
User                    Meta Provider              File Provider
 │                         │                          │
 ├─ file upload ──────────────────────────────────────►│ (object store)
 │                         │                          │
 ├─ cron create ──────────►│                          │
 │   (--object, --schedule)│                          │
 │                         ├─ Deploy(object, path, ──►│
 │                         │   mode, content_type)    │
 │                         │                          ├─ fetch from obj store
 │                         │                          ├─ render template (if applicable)
 │                         │                          ├─ SHA check (idempotent)
 │                         │                          ├─ write to disk
 │                         │                          ├─ update file-state KV
 │                         │◄─ DeployResult ──────────┤
 │◄─ CronCreateResponse ──┤                          │
```

### Examples Across Domains

| Meta Provider | Object Content          | Deploy Path                       | Mode |
| ------------- | ----------------------- | --------------------------------- | ---- |
| cron (sched)  | cron.d formatted line   | `/etc/cron.d/{name}`              | 0644 |
| cron (intv)   | shell script            | `/etc/cron.{interval}/{name}`     | 0755 |
| systemd       | unit file               | `/etc/systemd/system/{name}`      | 0644 |
| sysctl        | sysctl conf             | `/etc/sysctl.d/{name}.conf`       | 0644 |
| apt sources   | repo entry              | `/etc/apt/sources.list.d/{name}`  | 0644 |

All meta providers follow the same flow: user uploads content → meta provider
determines path + permissions + validation → `fileProvider.Deploy()` → SHA
tracked, idempotent, no magic headers.

### FileDeployer Interface

Meta providers depend on a narrow interface, not the full `file.Provider`:

```go
// FileDeployer is the narrow interface for providers that deploy
// files to well-known paths. Cron, systemd, sysctl, etc.
type FileDeployer interface {
    Deploy(ctx context.Context, req DeployRequest) (*DeployResult, error)
    Undeploy(ctx context.Context, req UndeployRequest) (*UndeployResult, error)
}
```

The existing `file.Service` satisfies `FileDeployer` automatically since it
already has `Deploy`. We add `Undeploy` as part of this work.

### Template Rendering

The file provider already supports Go `text/template` rendering when
`ContentType: "template"`. Templates have access to:

```go
type TemplateContext struct {
    Facts    map[string]any  // agent facts (arch, kernel, OS, etc.)
    Vars     map[string]any  // user-supplied variables
    Hostname string          // agent hostname
}
```

Meta providers pass `ContentType` and `Vars` through to `Deploy()`. A cron
script template can reference `{{ .Hostname }}`, `{{ .Facts.os_family }}`, or
user-supplied `{{ .Vars.region }}`. The same uploaded template renders
differently per host.

## File Provider Changes

### Undeploy Method

Removes a deployed file from disk. The object stays in the object store. The
file-state KV entry is updated to record the undeploy (not deleted — it serves
as an audit trail).

```go
type UndeployRequest struct {
    Path string `json:"path"`
}

type UndeployResult struct {
    Changed bool   `json:"changed"`
    Path    string `json:"path"`
}
```

Behavior:

- If file exists on disk: remove it, update file-state KV, `Changed: true`
- If file does not exist: no-op, `Changed: false`
- Object store entry is untouched
- `client file list` still shows the object

### Undeploy API Endpoint

```
DELETE /node/{hostname}/file/deploy/{name}
```

Removes the deployed file from disk on the target node. The object stays in the
store for redeployment or audit purposes. This is the inverse of
`POST /node/{hostname}/file/deploy`.

### Protected Objects

System-managed templates ship with osapi and cannot be deleted by users. These
are templates that meta providers reference (e.g., a standard systemd unit
template).

**Storage:** Objects in the NATS object store with a `system/` name prefix are
protected. Convention-based, no metadata changes needed.

```
system/systemd-unit.tmpl      → protected (cannot delete)
system/sysctl-conf.tmpl       → protected (cannot delete)
backup.sh                     → user-managed (deletable)
my-nginx.conf                 → user-managed (deletable)
```

**Enforcement:** The `file delete` handler checks if the object name starts with
`system/` and returns 403 if so.

**Seeding:** The agent seeds system templates into the object store on startup
(idempotent — skip if already present). Templates are embedded in the binary
via `go:embed`.

**Listing:** `client file list` shows both system and user objects. A `source`
column indicates `system` vs `user`.

## Cron Provider Refactor

### API Changes

The `command` field is removed. A new `object` field references an uploaded file
in the object store. The cron provider deploys the object to the correct path
with the correct permissions.

**CronCreateRequest:**

```yaml
CronCreateRequest:
  type: object
  required:
    - name
    - object
  properties:
    name:
      type: string
      description: >
        Name for the cron entry. Used as the filename under
        /etc/cron.d/ or /etc/cron.{interval}/.
    object:
      type: string
      description: >
        Name of the uploaded file in the object store to deploy
        as the cron entry content.
    schedule:
      type: string
      description: >
        Cron schedule expression (e.g., "*/5 * * * *"). Mutually
        exclusive with interval.
    interval:
      type: string
      description: >
        Periodic interval (hourly, daily, weekly, monthly). Mutually
        exclusive with schedule.
      enum: [hourly, daily, weekly, monthly]
    user:
      type: string
      description: >
        User to run the command as. Only applies to cron.d entries.
    content_type:
      type: string
      description: >
        "raw" or "template". When "template", the file content is
        rendered through Go's text/template engine with facts and vars.
      enum: [raw, template]
      default: raw
    vars:
      type: object
      description: >
        Template variables. Only used when content_type is "template".
```

**CronUpdateRequest:**

```yaml
CronUpdateRequest:
  type: object
  properties:
    object:
      type: string
      description: >
        New object to deploy (redeploy with updated content).
    schedule:
      type: string
    user:
      type: string
    content_type:
      type: string
      enum: [raw, template]
    vars:
      type: object
```

### Provider Changes

The Debian cron provider takes a `FileDeployer` dependency:

```go
type Debian struct {
    logger       *slog.Logger
    fs           afero.Fs
    fileDeployer file.FileDeployer
}
```

**Create:**

1. Validate name and schedule/interval
2. Check uniqueness across all cron directories
3. Determine path and mode:
   - Schedule → `/etc/cron.d/{name}`, 0644
   - Interval → `/etc/cron.{interval}/{name}`, 0755
4. Call `fileDeployer.Deploy(ctx, file.DeployRequest{
       ObjectName: entry.Object,
       Path: path,
       Mode: mode,
       ContentType: entry.ContentType,
       Vars: entry.Vars,
   })`
5. Return `CreateResult{Changed: result.Changed}`

**Update:**

1. Validate name, find existing entry path
2. Call `fileDeployer.Deploy()` with new object/vars — idempotent, skips if
   SHA unchanged
3. Return `UpdateResult{Changed: result.Changed}`

**Delete:**

1. Validate name, find existing entry path
2. Call `fileDeployer.Undeploy(ctx, file.UndeployRequest{Path: path})`
3. Return `DeleteResult{Changed: result.Changed}`

**List:**

1. Scan `/etc/cron.d/` and `/etc/cron.{interval}/` directories
2. For each file, compute the file-state KV key and check if it has a state
   entry — if yes, it is managed by osapi
3. Return entries with metadata from the file-state KV (object name, SHA, etc.)
4. No `# Managed by osapi` header — the file-state KV is the source of truth

**Get:**

1. Compute file-state KV key for the expected path
2. Look up state entry
3. Read file from disk for current content
4. Return entry with metadata

### Removed

- `buildFileContent()` — content comes from the uploaded object
- `# Managed by osapi` header — file-state KV is the source of truth
- `command` field — replaced by `object` reference
- Direct `afero.WriteFile` calls — replaced by `fileDeployer.Deploy()`

## Agent Wiring Changes

The cron provider factory needs the file provider:

```go
// factory.go
func (f *ProviderFactory) CreateProviders() (...) {
    // ...
    var cronProvider cronProv.Provider
    switch plat {
    case "debian":
        cronProvider = cronProv.NewDebianProvider(
            f.logger, f.appFs, fileProvider)
    // ...
    }
}
```

The cron provider must be created after the file provider. The factory already
creates the file provider first.

## SDK Changes

### Client

- `CronCreateOpts`: remove `Command`, add `Object`, `ContentType`, `Vars`
- `CronUpdateOpts`: remove `Command`, add `Object`, `ContentType`, `Vars`
- Add `FileUndeploy(ctx, hostname, name)` method to `FileService`
- Add `source` field to file list results (system vs user)

### CLI

- `client node schedule cron create`: remove `--command`, add `--object`,
  `--content-type`, `--vars`
- `client node schedule cron update`: remove `--command`, add `--object`,
  `--content-type`, `--vars`
- `client node file undeploy`: new command to remove deployed file from disk
- `client node file list`: add SOURCE column (system/user)

## Scope Summary

| Area              | Change                                                    |
| ----------------- | --------------------------------------------------------- |
| File provider     | Add `Undeploy` method, `FileDeployer` interface           |
| File API          | Add undeploy endpoint                                     |
| File handler      | Protected object check on delete, undeploy handler        |
| File SDK/CLI      | `undeploy` command, `source` column in list               |
| Cron provider     | Refactor to use `FileDeployer`, remove direct file writes |
| Cron API          | `command` → `object`, add `content_type`/`vars`           |
| Cron handler      | Pass through new fields                                   |
| Cron SDK/CLI      | Update opts and flags                                     |
| Agent wiring      | Pass file provider to cron provider                       |
| System templates  | `go:embed` + seeding on startup                           |
| Tests             | All of the above                                          |
| Docs              | Feature pages, CLI docs, API docs                         |
