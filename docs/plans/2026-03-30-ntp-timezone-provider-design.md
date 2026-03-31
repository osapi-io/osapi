# NTP + Timezone Provider Design

## Overview

Add NTP server management and timezone configuration to OSAPI. Two
separate providers under `provider/node/`, two separate API packages
under `api/node/`. NTP manages chrony server configuration via
drop-in files. Timezone reads and sets the system timezone via
timedatectl.

## NTP Provider

### Architecture

Direct provider at `internal/provider/node/ntp/`. Manages a chrony
drop-in config file at `/etc/chrony/sources.d/osapi.sources`. Reads
sync status and configured sources via `chronyc` commands. Applies
changes via `chronyc reload sources`.

- **Category**: `node`
- **Path prefix**: `/node/{hostname}/ntp`
- **Permissions**: `ntp:read`, `ntp:write`
- **Provider type**: direct (writes config, runs commands)

### Provider Interface

```go
type Provider interface {
    Get(ctx context.Context) (*Status, error)
    Create(ctx context.Context, config Config) (*CreateResult, error)
    Update(ctx context.Context, config Config) (*UpdateResult, error)
    Delete(ctx context.Context) (*DeleteResult, error)
}
```

### Data Types

```go
type Config struct {
    Servers []string `json:"servers"`
}

type Status struct {
    Synchronized  bool     `json:"synchronized"`
    Stratum       int      `json:"stratum,omitempty"`
    Offset        string   `json:"offset,omitempty"`
    CurrentSource string   `json:"current_source,omitempty"`
    Servers       []string `json:"servers,omitempty"`
}

type CreateResult struct {
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}

type UpdateResult struct {
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}

type DeleteResult struct {
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}
```

### Debian Implementation

**Config file**: `/etc/chrony/sources.d/osapi.sources`

Content format (one server per line):
```
server 0.pool.ntp.org iburst
server 1.pool.ntp.org iburst
server time.google.com iburst
```

**Operations:**

- **Get**: Parse `chronyc tracking` for sync state (synchronized,
  stratum, offset, current source). Parse `chronyc sources` for
  configured server list. Always succeeds — returns current state
  regardless of whether osapi manages the config.
- **Create**: Write the drop-in file with the server list. Fail if
  the osapi drop-in already exists. Run `chronyc reload sources`.
- **Update**: Overwrite the drop-in file. Fail if the osapi drop-in
  does not exist. Idempotent — compare content, skip write if
  unchanged (`Changed: false`). Run `chronyc reload sources` if
  changed.
- **Delete**: Remove the drop-in file. Fail if not found. Run
  `chronyc reload sources` to revert to default sources.

**Idempotency**: SHA-based comparison of generated file content,
same approach as sysctl provider.

### Container Behavior

No `DebianDocker` variant needed. NTP is a host-level concern —
containers inherit the host's time. If the agent runs in a
container, chronyc is unlikely to be available and operations
return the standard error.

### API Endpoints

| Method   | Path                    | Permission  | Description                |
| -------- | ----------------------- | ----------- | -------------------------- |
| `GET`    | `/node/{hostname}/ntp`  | `ntp:read`  | Get sync status + servers  |
| `POST`   | `/node/{hostname}/ntp`  | `ntp:write` | Create managed NTP config  |
| `PUT`    | `/node/{hostname}/ntp`  | `ntp:write` | Update server list         |
| `DELETE` | `/node/{hostname}/ntp`  | `ntp:write` | Remove managed config      |

All endpoints support broadcast targeting.

### Response Shape

GET response:
```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "synchronized": true,
    "stratum": 2,
    "offset": "+0.003s",
    "current_source": "0.pool.ntp.org",
    "servers": ["0.pool.ntp.org", "1.pool.ntp.org"]
  }]
}
```

POST/PUT/DELETE response:
```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "changed": true
  }]
}
```

POST/PUT request body:
```json
{
  "servers": ["0.pool.ntp.org", "1.pool.ntp.org", "time.google.com"]
}
```

---

## Timezone Provider

### Architecture

Direct provider at `internal/provider/node/timezone/`. Reads and
sets the system timezone via `timedatectl`. No config files — this
is a direct system call.

- **Category**: `node`
- **Path prefix**: `/node/{hostname}/timezone`
- **Permissions**: `timezone:read`, `timezone:write`
- **Provider type**: direct

### Provider Interface

```go
type Provider interface {
    Get(ctx context.Context) (*Info, error)
    Update(ctx context.Context, timezone string) (*UpdateResult, error)
}
```

No Create or Delete — timezone always exists on the system. Only
GET (read) and PUT (update).

### Data Types

```go
type Info struct {
    Timezone  string `json:"timezone"`
    UTCOffset string `json:"utc_offset,omitempty"`
}

type UpdateResult struct {
    Timezone string `json:"timezone"`
    Changed  bool   `json:"changed"`
    Error    string `json:"error,omitempty"`
}
```

### Debian Implementation

- **Get**: Run `timedatectl show -p Timezone --value` for timezone
  name. Run `date +%:z` or parse timedatectl output for UTC offset.
- **Update**: Run `timedatectl set-timezone <tz>`. Idempotent —
  read current timezone first, skip if already set
  (`Changed: false`). Validate timezone name against
  `/usr/share/zoneinfo/` or `timedatectl list-timezones`.

### Container Behavior

No `DebianDocker` variant needed. Containers inherit the host
timezone. If the agent runs in a container, `timedatectl` may not
be available — operations return the standard error.

### API Endpoints

| Method | Path                          | Permission       | Description       |
| ------ | ----------------------------- | ---------------- | ----------------- |
| `GET`  | `/node/{hostname}/timezone`   | `timezone:read`  | Get timezone      |
| `PUT`  | `/node/{hostname}/timezone`   | `timezone:write` | Set timezone      |

All endpoints support broadcast targeting.

### Response Shape

GET response:
```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "timezone": "America/New_York",
    "utc_offset": "-04:00"
  }]
}
```

PUT request body:
```json
{
  "timezone": "America/New_York"
}
```

PUT response:
```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "timezone": "America/New_York",
    "changed": true
  }]
}
```

---

## Platform Implementations

| Provider | Debian         | Darwin         | Linux          |
| -------- | -------------- | -------------- | -------------- |
| NTP      | chronyc        | ErrUnsupported | ErrUnsupported |
| Timezone | timedatectl    | ErrUnsupported | ErrUnsupported |

---

## Files to Create/Modify

### New Files

```
internal/provider/node/ntp/
  types.go
  debian.go
  darwin.go
  linux.go
  mocks/generate.go

internal/provider/node/timezone/
  types.go
  debian.go
  darwin.go
  linux.go
  mocks/generate.go

internal/agent/processor_ntp.go
internal/agent/processor_timezone.go

internal/controller/api/node/ntp/
  gen/ (api.yaml, cfg.yaml, generate.go)
  types.go
  ntp.go
  ntp_get.go
  ntp_create.go
  ntp_update.go
  ntp_delete.go
  handler.go
  validate.go
  *_public_test.go

internal/controller/api/node/timezone/
  gen/ (api.yaml, cfg.yaml, generate.go)
  types.go
  timezone.go
  timezone_get.go
  timezone_update.go
  handler.go
  validate.go
  *_public_test.go

pkg/sdk/client/
  ntp.go
  ntp_types.go
  timezone.go
  timezone_types.go
  *_public_test.go

cmd/
  client_node_ntp.go
  client_node_ntp_get.go
  client_node_ntp_create.go
  client_node_ntp_update.go
  client_node_ntp_delete.go
  client_node_timezone.go
  client_node_timezone_get.go
  client_node_timezone_update.go

examples/sdk/client/ntp.go
examples/sdk/client/timezone.go

test/integration/ntp_test.go
test/integration/timezone_test.go

docs/docs/sidebar/features/ntp.md
docs/docs/sidebar/features/timezone.md
docs/docs/sidebar/usage/cli/client/node/ntp/...
docs/docs/sidebar/usage/cli/client/node/timezone/...
docs/docs/sidebar/sdk/client/ntp.md
docs/docs/sidebar/sdk/client/timezone.md
```

### Modified Files

```
pkg/sdk/client/operations.go        — Add OpNtp*, OpTimezone*
pkg/sdk/client/permissions.go       — Add PermNtp*, PermTimezone*
pkg/sdk/client/osapi.go             — Wire services
internal/job/types.go                — Re-export operations
internal/authtoken/permissions.go    — Re-export + add to roles
internal/agent/processor.go          — Add ntp/timezone to node processor
cmd/agent_setup.go                   — Create + register providers
cmd/controller_setup.go              — Register handlers
docs/docusaurus.config.ts            — Add to Features navbar
docs/docs/sidebar/usage/configuration.md — Add permissions
docs/docs/sidebar/features/authentication.md — Add permissions
docs/docs/sidebar/architecture/api-guidelines.md — Add endpoints
docs/docs/sidebar/architecture/architecture.md — Add feature links
CLAUDE.md                            — Update provider list
```
