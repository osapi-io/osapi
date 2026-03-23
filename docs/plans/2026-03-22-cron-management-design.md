# Cron Drop-in Management Design

## Goal

Add cron drop-in file management (`/etc/cron.d/`) to OSAPI. This is the first
provider under the `scheduled/` domain. Crontab (user crontabs) and systemd
timer management will follow as separate providers.

## Provider

**Location:** `internal/provider/scheduled/cron/`

**Interface:**

```go
type Provider interface {
    List() ([]CronEntry, error)
    Get(name string) (*CronEntry, error)
    Create(entry CronEntry) (*CreateResult, error)
    Update(entry CronEntry) (*UpdateResult, error)
    Delete(name string) (*DeleteResult, error)
}
```

**CronEntry:**

```go
type CronEntry struct {
    Name     string `json:"name"`
    Schedule string `json:"schedule"`
    User     string `json:"user"`
    Command  string `json:"command"`
}
```

**Result types** include `Changed bool` and `Error string` fields per
convention.

**Debian provider:** Reads and writes `/etc/cron.d/{name}` files using
`afero.Fs`. Each file contains:

```
# Managed by osapi
SCHEDULE USER COMMAND
```

File permissions: 0644 (standard for `/etc/cron.d/` files). The name is
sanitized to prevent path traversal. Names must be alphanumeric with hyphens
and underscores only.

**Darwin provider:** Returns `provider.ErrUnsupported` for all operations.
Jobs targeting macOS agents get `StatusSkipped`.

**Linux stub:** Returns `provider.ErrUnsupported`.

## API

**Path:** `/node/{hostname}/schedule/cron` and
`/node/{hostname}/schedule/cron/{name}`

**Endpoints:**

| Method | Path | Operation | Permission |
|--------|------|-----------|------------|
| GET | `/node/{hostname}/schedule/cron` | `cron.list` | `cron:read` |
| GET | `/node/{hostname}/schedule/cron/{name}` | `cron.get` | `cron:read` |
| POST | `/node/{hostname}/schedule/cron` | `cron.create` | `cron:write` |
| PUT | `/node/{hostname}/schedule/cron/{name}` | `cron.update` | `cron:write` |
| DELETE | `/node/{hostname}/schedule/cron/{name}` | `cron.delete` | `cron:write` |

**OpenAPI spec:** `internal/controller/api/schedule/gen/api.yaml`

**Validation (via `x-oapi-codegen-extra-tags`):**

- `name` — required, alphanum with hyphens/underscores
- `schedule` — required (cron expression, validated by provider)
- `command` — required
- `user` — optional, defaults to `root`

**Response format:** Same collection/result pattern as other domains.
List returns `CronCollectionResponse` with `results` array.

## Job Routing

- Query: `cron.list`, `cron.get`
- Modify: `cron.create`, `cron.update`, `cron.delete`

**Job category:** `schedule`
**Job operations:** `cron.list`, `cron.get`, `cron.create`, `cron.update`,
`cron.delete`

## SDK

**New operations in `pkg/sdk/client/operations.go`:**

```go
OpCronList   JobOperation = "cron.list"
OpCronGet    JobOperation = "cron.get"
OpCronCreate JobOperation = "cron.create"
OpCronUpdate JobOperation = "cron.update"
OpCronDelete JobOperation = "cron.delete"
```

**New permissions in `pkg/sdk/client/permissions.go`:**

```go
PermCronRead  Permission = "cron:read"
PermCronWrite Permission = "cron:write"
```

Add `cron:read` and `cron:write` to the `admin` and `write` default roles.
Add `cron:read` to the `read` role.

**New `CronService`** on the SDK client with typed result types in
`pkg/sdk/client/schedule.go` and `pkg/sdk/client/schedule_types.go`.

## CLI

```
osapi client node schedule cron list   --hostname web-01
osapi client node schedule cron get    --hostname web-01 --name backup
osapi client node schedule cron create --hostname web-01 --name backup \
    --schedule "0 2 * * *" --command "/usr/local/bin/backup.sh" --user root
osapi client node schedule cron update --hostname web-01 --name backup \
    --schedule "0 3 * * *"
osapi client node schedule cron delete --hostname web-01 --name backup
```

All commands support `--json` for raw output.

## Additional Changes

- Move `internal/provider/process/` to `internal/provider/node/process/`
- Add cron provider to agent factory with platform switch
- Wire schedule handler in controller setup
- Update CLAUDE.md with new domain
- Documentation: feature page, CLI reference, API reference, config reference
