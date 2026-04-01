# Package Management Provider Design

## Overview

Add package management to OSAPI. List installed packages, get details, install,
remove, refresh package sources, and list available updates. Uses `apt-get` and
`dpkg-query` via `exec.Manager`. Provider package is `apt` (the tool), API path
is `/node/{hostname}/package` (the concept).

## Architecture

Direct provider at `internal/provider/node/apt/`.

- **Category**: `node`
- **Path prefix**: `/node/{hostname}/package`
- **Permissions**: `package:read`, `package:write`
- **Provider type**: direct (exec.Manager)

## Provider Interface

```go
type Provider interface {
    List(ctx context.Context) ([]Package, error)
    Get(ctx context.Context, name string) (*Package, error)
    Install(ctx context.Context, name string) (*Result, error)
    Remove(ctx context.Context, name string) (*Result, error)
    Update(ctx context.Context) (*Result, error)
    ListUpdates(ctx context.Context) ([]Update, error)
}
```

## Data Types

```go
type Package struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    Description string `json:"description,omitempty"`
    Status      string `json:"status"`
    Size        int64  `json:"size,omitempty"`
}

type Update struct {
    Name           string `json:"name"`
    CurrentVersion string `json:"current_version"`
    NewVersion     string `json:"new_version"`
}

type Result struct {
    Name    string `json:"name,omitempty"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}
```

## Debian Implementation

- **List**: `dpkg-query -W -f` with format string to get name, version,
  description, status, installed size. Parse tab-separated output. Filter for
  `install ok installed` status.
- **Get**: same query filtered to one package. Error if not found.
- **Install**: `apt-get install -y <name>`. Return Changed: true.
- **Remove**: `apt-get remove -y <name>`. Return Changed: true.
- **Update**: `apt-get update`. Refreshes package index. Return Changed: true.
- **ListUpdates**: `apt list --upgradable`. Parse output for package name,
  current version, and new version.

## Platform Implementations

| Platform | Implementation       |
| -------- | -------------------- |
| Debian   | apt-get / dpkg-query |
| Darwin   | ErrUnsupported       |
| Linux    | ErrUnsupported       |

## Container Behavior

Return `ErrUnsupported` in containers.

## API Endpoints

| Method   | Path                              | Permission      | Description             |
| -------- | --------------------------------- | --------------- | ----------------------- |
| `GET`    | `/node/{hostname}/package`        | `package:read`  | List installed packages |
| `GET`    | `/node/{hostname}/package/{name}` | `package:read`  | Get package details     |
| `POST`   | `/node/{hostname}/package`        | `package:write` | Install a package       |
| `DELETE` | `/node/{hostname}/package/{name}` | `package:write` | Remove a package        |
| `POST`   | `/node/{hostname}/package/update` | `package:write` | Refresh package sources |
| `GET`    | `/node/{hostname}/package/update` | `package:read`  | List available updates  |

All endpoints support broadcast targeting.

### POST Install Request Body

```json
{
  "name": "nginx"
}
```

Name is required.

### Response Shapes

List response:

```json
{
  "job_id": "...",
  "results": [
    {
      "hostname": "web-01",
      "status": "ok",
      "packages": [
        {
          "name": "nginx",
          "version": "1.24.0-1",
          "status": "installed",
          "size": 1234567
        }
      ]
    }
  ]
}
```

Install/Remove/Update response:

```json
{
  "job_id": "...",
  "results": [
    {
      "hostname": "web-01",
      "status": "ok",
      "name": "nginx",
      "changed": true
    }
  ]
}
```

List updates response:

```json
{
  "job_id": "...",
  "results": [
    {
      "hostname": "web-01",
      "status": "ok",
      "updates": [
        {
          "name": "openssl",
          "current_version": "3.0.11-1",
          "new_version": "3.0.13-1"
        }
      ]
    }
  ]
}
```

## SDK

```go
client.Package.List(ctx, host)
client.Package.Get(ctx, host, name)
client.Package.Install(ctx, host, name)
client.Package.Remove(ctx, host, name)
client.Package.Update(ctx, host)
client.Package.ListUpdates(ctx, host)
```

## Permissions

- `package:read` — list, get, list updates. Added to admin, write, and read
  roles.
- `package:write` — install, remove, update sources. Added to admin and write
  roles.
