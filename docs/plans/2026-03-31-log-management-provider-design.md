# Log Viewing Provider Design

## Overview

Add log viewing to OSAPI. Query systemd journal entries with
optional filtering by lines, time range, and priority. Read-only
— no write operations. Uses `journalctl --output=json` for
structured parsing.

## Architecture

Direct provider at `internal/provider/node/log/`.

- **Category**: `node`
- **Path prefix**: `/node/{hostname}/log`
- **Permissions**: `log:read`
- **Provider type**: direct (exec.Manager)

## Provider Interface

```go
type Provider interface {
    Query(ctx context.Context, opts QueryOpts) ([]Entry, error)
    QueryUnit(ctx context.Context, unit string, opts QueryOpts) ([]Entry, error)
}
```

## Data Types

```go
type QueryOpts struct {
    Lines    int    `json:"lines,omitempty"`
    Since    string `json:"since,omitempty"`
    Priority string `json:"priority,omitempty"`
}

type Entry struct {
    Timestamp string `json:"timestamp"`
    Unit      string `json:"unit,omitempty"`
    Priority  string `json:"priority"`
    Message   string `json:"message"`
    PID       int    `json:"pid,omitempty"`
    Hostname  string `json:"hostname,omitempty"`
}
```

## Debian Implementation

- **Query**: run `journalctl --output=json -n <lines>` with
  optional `--since=<since>` and `--priority=<priority>`. Parse
  JSON lines output — each line is a JSON object with fields
  `__REALTIME_TIMESTAMP`, `SYSLOG_IDENTIFIER`, `PRIORITY`,
  `MESSAGE`, `_PID`, `_HOSTNAME`.
- **QueryUnit**: same but with `-u <unit>` flag.

Default `lines` is 100 if not specified. `since` uses journalctl
format (e.g., `"1 hour ago"`, `"2026-03-31"`). `priority` uses
journalctl levels (0-7 or names like `err`, `warning`).

## Platform Implementations

| Platform | Implementation             |
| -------- | -------------------------- |
| Debian   | journalctl --output=json   |
| Darwin   | ErrUnsupported             |
| Linux    | ErrUnsupported             |

## Container Behavior

Return `ErrUnsupported` in containers — `journalctl` requires
systemd which isn't available in containers.

## API Endpoints

| Method | Path                              | Permission | Description                 |
| ------ | --------------------------------- | ---------- | --------------------------- |
| `GET`  | `/node/{hostname}/log`            | `log:read` | Query journal entries       |
| `GET`  | `/node/{hostname}/log/unit/{name}`| `log:read` | Query entries for a unit    |

All endpoints support broadcast targeting.

### Query Parameters

| Param      | Type    | Default | Description                                |
| ---------- | ------- | ------- | ------------------------------------------ |
| `lines`    | integer | 100     | Number of entries to return                |
| `since`    | string  |         | Time filter (e.g., "1 hour ago")           |
| `priority` | string  |         | Minimum priority (emerg..debug or 0-7)     |

### Response Shape

```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "entries": [
      {
        "timestamp": "2026-03-31T22:30:45.123Z",
        "unit": "nginx.service",
        "priority": "info",
        "message": "Started nginx",
        "pid": 1234,
        "hostname": "web-01"
      }
    ]
  }]
}
```

## SDK

```go
client.Log.Query(ctx, host, opts)
client.Log.QueryUnit(ctx, host, unit, opts)
```

`LogQueryOpts` struct with optional `Lines`, `Since`, `Priority`.

## Permissions

- `log:read` — query journal entries. Added to admin, write, and
  read roles.
