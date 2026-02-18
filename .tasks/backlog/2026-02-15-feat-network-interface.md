---
title: Network interface configuration
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add network interface management endpoints. Currently only DNS and ping
are supported. A real appliance needs IP address assignment, interface
up/down, and routing.

## API Endpoints

```
GET    /network/interface              - List network interfaces
GET    /network/interface/{name}       - Get interface details
PUT    /network/interface/{name}       - Update interface config (IP, mask)
POST   /network/interface/{name}/up    - Bring interface up
POST   /network/interface/{name}/down  - Bring interface down

GET    /network/route                  - List routing table
POST   /network/route                  - Add route
DELETE /network/route/{id}             - Delete route
```

## Operations

- `network.interface.list.get`, `network.interface.status.get` (query)
- `network.interface.update` (modify)
- `network.interface.up.execute`, `network.interface.down.execute` (modify)
- `network.route.list.get` (query)
- `network.route.create`, `network.route.delete` (modify)

## Provider

- `internal/provider/network/interface/`
- Parse `/sys/class/net/`, use `ip` command
- Implementations: `netplan_provider.go` (Ubuntu),
  `linux_provider.go` (generic)
- Return type: `InterfaceInfo` with name, MAC, IPs, MTU, state,
  speed, duplex, stats (rx/tx bytes)

## Notes

- Interface changes can disconnect the management network — warn user
- Scopes: `network:read`, `network:write`
- Follows the existing `/network/` path convention
