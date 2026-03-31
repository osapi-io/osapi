# Power Management Provider Design

## Overview

Add power management (reboot/shutdown) to OSAPI. Two action
operations on a direct provider — no persistent resources, no file
management. The provider runs the `shutdown` command with a minimum
5-second implicit delay so the agent can complete its job response
lifecycle before the system goes down.

## Architecture

Direct provider at `internal/provider/node/power/`. Action
operations only (no CRUD).

- **Category**: `node`
- **Path prefix**: `/node/{hostname}/power`
- **Permissions**: `power:execute`
- **Provider type**: direct

## Provider Interface

```go
type Provider interface {
    Reboot(ctx context.Context, opts Opts) (*Result, error)
    Shutdown(ctx context.Context, opts Opts) (*Result, error)
}
```

## Data Types

```go
type Opts struct {
    Delay   int    `json:"delay,omitempty"`
    Message string `json:"message,omitempty"`
}

type Result struct {
    Action  string `json:"action"`
    Delay   int    `json:"delay"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}
```

- `Delay` — seconds before the action. Minimum 5 seconds enforced
  by the provider regardless of what the user requests. This gives
  the agent time to write the job result, send the response, and
  run graceful shutdown.
- `Message` — optional human-readable reason. Logged by the agent
  before executing.

## Debian Implementation

- **Reboot**: run `shutdown -r` with the computed delay
- **Shutdown**: run `shutdown -h` with the computed delay
- Actual delay = `max(userDelay, 5)` seconds
- Provider returns immediately with `changed: true` and the actual
  delay applied
- The agent completes its KV write and response lifecycle before
  the system goes down
- Use `exec.Manager` for running the shutdown command

## Platform Implementations

| Platform | Implementation           |
| -------- | ------------------------ |
| Debian   | `shutdown -r` / `-h`    |
| Darwin   | ErrUnsupported           |
| Linux    | ErrUnsupported           |

No container variant needed — power management doesn't make sense
inside a container.

## API Endpoints

| Method | Path                               | Permission       | Description  |
| ------ | ---------------------------------- | ---------------- | ------------ |
| `POST` | `/node/{hostname}/power/reboot`    | `power:execute`  | Reboot node  |
| `POST` | `/node/{hostname}/power/shutdown`  | `power:execute`  | Shutdown node |

All endpoints support broadcast targeting.

### Request Body (optional)

```json
{
  "delay": 60,
  "message": "Scheduled maintenance"
}
```

Both fields optional. If omitted, immediate action (with the
5-second minimum implicit delay).

### Validation

- `delay`: integer, min 0, optional
- `message`: string, optional

### Response Shape

```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "action": "reboot",
    "delay": 60,
    "changed": true
  }]
}
```

## SDK

```go
client.Power.Reboot(ctx, host, opts)
client.Power.Shutdown(ctx, host, opts)
```

`PowerOpts` struct with optional `Delay` and `Message` fields.
Both methods return `*Response[Collection[PowerResult]]`.

## Permission

`power:execute` — action permission, same pattern as
`command:execute`. No read permission exists for this domain.

Added to admin role only (not write or read) — power operations
are destructive and should require explicit authorization.
