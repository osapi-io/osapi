# Service Management Provider Design

## Overview

Add systemd service management to OSAPI. List and inspect services, control them
(start/stop/restart/enable/disable), and manage custom unit files via Object
Store deployment. Hybrid provider — direct for control operations, meta for unit
file CRUD.

## Architecture

Hybrid provider at `internal/provider/node/service/`.

- **Category**: `node`
- **Path prefix**: `/node/{hostname}/service`
- **Permissions**: `service:read`, `service:write`
- **Provider type**: hybrid (exec.Manager + file.Deployer)

## Provider Interface

```go
type Provider interface {
    // Read
    List(ctx context.Context) ([]Info, error)
    Get(ctx context.Context, name string) (*Info, error)
    // Unit file CRUD (meta provider pattern)
    Create(ctx context.Context, entry Entry) (*CreateResult, error)
    Update(ctx context.Context, entry Entry) (*UpdateResult, error)
    Delete(ctx context.Context, name string) (*DeleteResult, error)
    // Control actions (direct provider pattern)
    Start(ctx context.Context, name string) (*ActionResult, error)
    Stop(ctx context.Context, name string) (*ActionResult, error)
    Restart(ctx context.Context, name string) (*ActionResult, error)
    Enable(ctx context.Context, name string) (*ActionResult, error)
    Disable(ctx context.Context, name string) (*ActionResult, error)
}
```

## Data Types

```go
type Info struct {
    Name        string `json:"name"`
    Status      string `json:"status"`
    Enabled     bool   `json:"enabled"`
    Description string `json:"description,omitempty"`
    PID         int    `json:"pid,omitempty"`
}

type Entry struct {
    Name   string `json:"name"`
    Object string `json:"object,omitempty"`
}

type CreateResult struct {
    Name    string `json:"name"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}

type UpdateResult struct {
    Name    string `json:"name"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}

type DeleteResult struct {
    Name    string `json:"name"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}

type ActionResult struct {
    Name    string `json:"name"`
    Changed bool   `json:"changed"`
}
```

## Debian Implementation

The Debian struct needs:

- `provider.FactsAware` embedded
- `logger *slog.Logger`
- `fs avfs.VFS` — unit file existence checks
- `fileDeployer file.Deployer` — unit file deployment
- `stateKV jetstream.KeyValue` — managed file tracking
- `execManager exec.Manager` — systemctl commands
- `hostname string` — state key construction

### Read Operations

- **List**: Run `systemctl list-units --type=service --all --output=json`. Parse
  JSON output into `[]Info`. Each entry maps `ActiveState` to status and
  `UnitFileState` to enabled.
- **Get**: Run
  `systemctl show {name} --property=ActiveState,UnitFileState,Description,MainPID`.
  Parse key=value output into `Info`.

### Unit File CRUD (Meta Provider)

- **Create**: Deploy unit file from Object Store to
  `/etc/systemd/system/osapi-{name}.service` via `file.Deployer` with mode
  `0644`. Run `systemctl daemon-reload`. Fails if unit already exists.
- **Update**: Redeploy unit file to same path. `file.Deployer` compares SHA — if
  unchanged, returns `changed: false` and skips `daemon-reload`. If object not
  specified, preserve existing (read from state KV).
- **Delete**: Stop and disable the service first (best-effort), undeploy via
  `file.Deployer`, run `systemctl daemon-reload`.

### Control Actions (Direct Provider)

- **Start**: Check current state via `systemctl is-active {name}`. If already
  active, return `changed: false`. Otherwise run `systemctl start {name}`.
- **Stop**: Check if active. If already inactive, return `changed: false`.
  Otherwise run `systemctl stop {name}`.
- **Restart**: Always run `systemctl restart {name}`, return `changed: true`. No
  idempotency check — restart is always intentional.
- **Enable**: Check via `systemctl is-enabled {name}`. If already enabled,
  return `changed: false`. Otherwise run `systemctl enable {name}`.
- **Disable**: Check if enabled. If already disabled, return `changed: false`.
  Otherwise run `systemctl disable {name}`.

## Platform Implementations

| Platform | Implementation            |
| -------- | ------------------------- |
| Debian   | systemctl + file.Deployer |
| Darwin   | ErrUnsupported            |
| Linux    | ErrUnsupported            |

## Container Behavior

Return `ErrUnsupported` in containers — systemctl requires systemd which isn't
available in standard containers.

## API Endpoints

| Method   | Path                                      | Permission      | Description         |
| -------- | ----------------------------------------- | --------------- | ------------------- |
| `GET`    | `/node/{hostname}/service`                | `service:read`  | List all services   |
| `GET`    | `/node/{hostname}/service/{name}`         | `service:read`  | Get service details |
| `POST`   | `/node/{hostname}/service`                | `service:write` | Create unit file    |
| `PUT`    | `/node/{hostname}/service/{name}`         | `service:write` | Update unit file    |
| `DELETE` | `/node/{hostname}/service/{name}`         | `service:write` | Delete unit file    |
| `POST`   | `/node/{hostname}/service/{name}/start`   | `service:write` | Start service       |
| `POST`   | `/node/{hostname}/service/{name}/stop`    | `service:write` | Stop service        |
| `POST`   | `/node/{hostname}/service/{name}/restart` | `service:write` | Restart service     |
| `POST`   | `/node/{hostname}/service/{name}/enable`  | `service:write` | Enable at boot      |
| `POST`   | `/node/{hostname}/service/{name}/disable` | `service:write` | Disable at boot     |

All endpoints support broadcast targeting.

### POST Request Body (Create)

```json
{
  "name": "my-app",
  "object": "my-app-unit"
}
```

`object` references an existing Object Store upload containing the systemd unit
file content. The provider writes it to
`/etc/systemd/system/osapi-{name}.service`.

### PUT Request Body (Update)

```json
{
  "object": "my-app-unit-v2"
}
```

Name comes from path parameter. Object is the new unit file content from Object
Store.

### Response Shape (List)

```json
{
  "job_id": "...",
  "results": [
    {
      "hostname": "web-01",
      "status": "ok",
      "services": [
        {
          "name": "nginx.service",
          "status": "active",
          "enabled": true,
          "description": "A high performance web server",
          "pid": 1234
        }
      ]
    }
  ]
}
```

### Response Shape (Get)

```json
{
  "job_id": "...",
  "results": [
    {
      "hostname": "web-01",
      "status": "ok",
      "service": {
        "name": "nginx.service",
        "status": "active",
        "enabled": true,
        "description": "A high performance web server",
        "pid": 1234
      }
    }
  ]
}
```

### Response Shape (Actions + CRUD)

```json
{
  "job_id": "...",
  "results": [
    {
      "hostname": "web-01",
      "status": "ok",
      "name": "nginx.service",
      "changed": true
    }
  ]
}
```

## SDK

```go
client.Service.List(ctx, host)
client.Service.Get(ctx, host, name)
client.Service.Create(ctx, host, opts)
client.Service.Update(ctx, host, name, opts)
client.Service.Delete(ctx, host, name)
client.Service.Start(ctx, host, name)
client.Service.Stop(ctx, host, name)
client.Service.Restart(ctx, host, name)
client.Service.Enable(ctx, host, name)
client.Service.Disable(ctx, host, name)
```

`ServiceCreateOpts` with `Name` and `Object` fields. `ServiceUpdateOpts` with
`Object` field.

## Permissions

- `service:read` — list and get. Added to admin, write, and read roles.
- `service:write` — create, update, delete, start, stop, restart, enable,
  disable. Added to admin and write roles.
