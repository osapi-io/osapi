# Process Management Provider Design

## Overview

Add process management to OSAPI. List running processes, get
details by PID, and send signals (TERM, KILL, HUP, etc.). Direct
provider using gopsutil for process info and syscall.Kill for
signaling.

## Architecture

Direct provider at `internal/provider/node/process/`.

- **Category**: `node`
- **Path prefix**: `/node/{hostname}/process`
- **Permissions**: `process:read`, `process:execute`
- **Provider type**: direct (gopsutil + syscall)

## Provider Interface

```go
type Provider interface {
    List(ctx context.Context) ([]Info, error)
    Get(ctx context.Context, pid int) (*Info, error)
    Signal(ctx context.Context, pid int, signal string) (*SignalResult, error)
}
```

## Data Types

```go
type Info struct {
    PID        int     `json:"pid"`
    Name       string  `json:"name"`
    User       string  `json:"user"`
    State      string  `json:"state"`
    CPUPercent float64 `json:"cpu_percent"`
    MemPercent float32 `json:"mem_percent"`
    MemRSS     int64   `json:"mem_rss"`
    Command    string  `json:"command"`
    StartTime  string  `json:"start_time"`
}

type SignalResult struct {
    PID     int    `json:"pid"`
    Signal  string `json:"signal"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}
```

## Debian Implementation

- **List**: `gopsutil/process.Processes()` to get all PIDs, collect
  info per process (name, user, state, CPU%, mem%, RSS, command,
  start time). Return `[]Info`.
- **Get**: `gopsutil/process.NewProcess(pid)` and read info. Return
  error if PID doesn't exist.
- **Signal**: validate signal name against allowed set (TERM, KILL,
  HUP, INT, USR1, USR2). Convert to `syscall.Signal`. Call
  `syscall.Kill(pid, sig)`. Return `SignalResult{Changed: true}`.

## Platform Implementations

| Platform | Implementation              |
| -------- | --------------------------- |
| Debian   | gopsutil + syscall.Kill     |
| Darwin   | ErrUnsupported              |
| Linux    | ErrUnsupported              |

## Container Behavior

Add `platform.IsContainer()` check in `agent_setup.go`. Process
management in containers returns ErrUnsupported — process
management is the host's concern.

## API Endpoints

| Method | Path                                    | Permission        | Description         |
| ------ | --------------------------------------- | ----------------- | ------------------- |
| `GET`  | `/node/{hostname}/process`              | `process:read`    | List all processes  |
| `GET`  | `/node/{hostname}/process/{pid}`        | `process:read`    | Get process by PID  |
| `POST` | `/node/{hostname}/process/{pid}/signal` | `process:execute` | Send signal to PID  |

All endpoints support broadcast targeting.

### POST Request Body

```json
{
  "signal": "TERM"
}
```

Signal is required. Valid values: TERM, KILL, HUP, INT, USR1, USR2.
Validated via `x-oapi-codegen-extra-tags` with `oneof`.

### Response Shapes

List response:
```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "processes": [
      {"pid": 1, "name": "systemd", "user": "root", "state": "S",
       "cpu_percent": 0.1, "mem_percent": 0.5, "mem_rss": 12345678,
       "command": "/sbin/init", "start_time": "2026-03-30T10:00:00Z"}
    ]
  }]
}
```

Get response — same shape but `processes` has one entry.

Signal response:
```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "pid": 1234,
    "signal": "TERM",
    "changed": true
  }]
}
```

## SDK

```go
client.Process.List(ctx, host)
client.Process.Get(ctx, host, pid)
client.Process.Signal(ctx, host, pid, opts)
```

`ProcessSignalOpts` has one required field: `Signal string`.

## Permissions

- `process:read` — list and get. Added to admin, write, and read
  roles.
- `process:execute` — send signals. Added to admin role only
  (destructive operation).
