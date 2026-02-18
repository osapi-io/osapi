---
title: "Feature: Process management"
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add process listing and management. Useful for troubleshooting and
monitoring resource consumption on the appliance.

## API Endpoints

```
GET    /process              - List running processes (with filters)
GET    /process/{pid}        - Get process details
POST   /process/{pid}/signal - Send signal to process (kill, term, hup)
```

## Operations

- `process.list.get` (query)
- `process.status.get` (query)
- `process.signal.execute` (modify)

## Provider

- `internal/provider/system/process/`
- Use gopsutil or parse `/proc/` directly
- Return type: `ProcessInfo` with PID, name, user, CPU%, memory%,
  state, command line, start time
- Support filtering by user, name pattern, state
- Support sorting by CPU, memory, PID

## Notes

- Killing processes is privileged — require elevated scope
- Scopes: `process:read`, `process:write`
- Consider pagination for large process lists
- Top-N by CPU/memory is a common use case
