---
title: Power management (shutdown/reboot)
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add power management endpoints. The operation constants
`system.shutdown.execute` and `system.reboot.execute` already exist in
`types.go` but have no provider, worker dispatch, API endpoint, or CLI command
implementation.

## API Endpoints

Per the api-guidelines, power operations belong in their own path:

```
POST   /power/shutdown      - Shutdown the system (with optional delay)
POST   /power/reboot        - Reboot the system (with optional delay)
```

## Operations

- `system.shutdown.execute` (modify) — already defined in types.go
- `system.reboot.execute` (modify) — already defined in types.go

## Provider

- `internal/provider/node/power/`
- Implementation uses `shutdown` command via cmdexec
- Request body: optional `delay` (seconds), optional `message`

## Implementation Notes

- Constants already exist — need provider + worker dispatch + API + CLI
- These are destructive operations: require confirmation or elevated scope
- Scopes: `power:write` (no read scope needed)
- Worker should write status event before executing shutdown
- API should return 202 Accepted (async) since the system will go down
- Consider a scheduled shutdown with cancel capability
