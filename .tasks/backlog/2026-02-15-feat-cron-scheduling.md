---
title: Cron/scheduled task management
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add scheduled task management. Appliances need recurring maintenance
operations (log rotation, backups, health checks) managed via API.

## API Endpoints

```
GET    /schedule             - List scheduled tasks (cron + systemd timers)
GET    /schedule/{id}        - Get schedule details
POST   /schedule             - Create scheduled task
PUT    /schedule/{id}        - Update schedule
DELETE /schedule/{id}        - Delete scheduled task
```

## Operations

- `schedule.list.get`, `schedule.status.get` (query)
- `schedule.create`, `schedule.update`, `schedule.delete` (modify)

## Provider

- `internal/provider/system/schedule/`
- Parse crontab files and `systemd-timer` units
- Support cron expression syntax for scheduling
- Return type: `ScheduleInfo` with ID, expression, command, user,
  next run, last run, enabled

## Notes

- Architecture.md already lists "Scheduled Jobs" as a future enhancement
- Could also integrate with the NATS job system for scheduled job
  submission
- Scopes: `schedule:read`, `schedule:write`
