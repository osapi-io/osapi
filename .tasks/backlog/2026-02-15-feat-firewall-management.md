---
title: Firewall management (ufw/nftables)
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add firewall rule management endpoints. An appliance needs to control
network access for security hardening.

## API Endpoints

```
GET    /firewall/status     - Get firewall status (active/inactive)
POST   /firewall/enable     - Enable firewall
POST   /firewall/disable    - Disable firewall
GET    /firewall/rule        - List firewall rules
POST   /firewall/rule        - Add firewall rule
DELETE /firewall/rule/{id}   - Delete firewall rule
```

## Operations

- `firewall.status.get` (query)
- `firewall.rules.get` (query)
- `firewall.enable.execute`, `firewall.disable.execute` (modify)
- `firewall.rule.create`, `firewall.rule.delete` (modify)

## Provider

- `internal/provider/network/firewall/`
- Implementations: `ufw_provider.go` (Ubuntu), `nftables_provider.go`
- Rule model: direction (in/out), action (allow/deny), protocol
  (tcp/udp/any), port, source, destination

## Notes

- Firewall changes are sensitive — consider confirmation or dry-run mode
- Scopes: `firewall:read`, `firewall:write`
- Should not allow rules that would lock out the API port itself
