---
title: "Feature: NTP and time management"
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add time and NTP management. Accurate time is critical for an appliance
(logging, certificates, authentication). The api-guidelines already
identify `/ntp/` as a planned path.

## API Endpoints

```
GET    /ntp/status           - Get NTP sync status and current time
GET    /ntp/server           - List configured NTP servers
PUT    /ntp/server           - Update NTP server list
POST   /ntp/sync             - Force time synchronization

GET    /time/timezone        - Get current timezone
PUT    /time/timezone        - Set timezone
```

## Operations

- `ntp.status.get`, `ntp.servers.get` (query)
- `ntp.servers.update`, `ntp.sync.execute` (modify)
- `time.timezone.get` (query)
- `time.timezone.update` (modify)

## Provider

- `internal/provider/system/ntp/`
- Implementations: `chrony_provider.go`, `timesyncd_provider.go`
- Use `timedatectl`, `chronyc` command parsing
- Return types: `NTPStatus` with sync state, stratum, offset, server,
  last sync time

## Notes

- Referenced in api-guidelines as a planned functional area
- Scopes: `ntp:read`, `ntp:write`
- Timezone changes affect all services on the host
