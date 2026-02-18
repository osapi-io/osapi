---
title: Service management (systemctl)
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add service management endpoints for controlling systemd services. This
is fundamental to any OS appliance — the ability to start, stop, restart,
enable, and disable services, and query their status.

## API Endpoints

```
GET    /service              - List all services (with status filter)
GET    /service/{name}       - Get service details (status, enabled, logs)
POST   /service/{name}/start   - Start a service
POST   /service/{name}/stop    - Stop a service
POST   /service/{name}/restart - Restart a service
PUT    /service/{name}/enable  - Enable service at boot
PUT    /service/{name}/disable - Disable service at boot
```

## Operations

- `service.list.get` (query) — list services with status
- `service.status.get` (query) — get single service status
- `service.start.execute` (modify) — start service
- `service.stop.execute` (modify) — stop service
- `service.restart.execute` (modify) — restart service
- `service.enable.execute` (modify) — enable at boot
- `service.disable.execute` (modify) — disable at boot

## Provider

- `internal/provider/system/service/`
- Interface: `Provider` with `List()`, `Status(name)`, `Start(name)`,
  `Stop(name)`, `Restart(name)`, `Enable(name)`, `Disable(name)`
- Implementation: `systemd_provider.go` using `systemctl` via cmdexec
- Return type: `ServiceInfo` with name, status (active/inactive/failed),
  enabled (true/false), description, PID, memory usage

## Notes

- Filter by state (running, stopped, failed, enabled, disabled)
- Consider pagination for list endpoint
- Scopes: `service:read`, `service:write`
