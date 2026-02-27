---
title: System facts/inventory gathering
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add a comprehensive system facts endpoint. Ansible's `setup` module (fact
gathering) is run automatically on every play — it collects hardware, OS,
network, and storage facts into a single structured document. This is invaluable
for inventory and fleet management.

## API Endpoints

```
GET    /facts                - Get all system facts
GET    /facts/{category}     - Get facts by category (hardware, os,
                               network, storage)
```

## Response Structure

```json
{
  "hostname": "server-01",
  "fqdn": "server-01.example.com",
  "os": {
    "distribution": "Ubuntu",
    "version": "24.04",
    "kernel": "6.8.0-45-generic",
    "arch": "x86_64"
  },
  "hardware": {
    "cpu_count": 4,
    "cpu_model": "Intel Xeon E-2236",
    "memory_total_mb": 32768,
    "swap_total_mb": 4096
  },
  "network": {
    "interfaces": [...],
    "default_gateway": "192.168.1.1",
    "dns_servers": [...]
  },
  "storage": {
    "disks": [...],
    "mounts": [...]
  },
  "virtualization": {
    "type": "kvm",
    "role": "guest"
  },
  "python_version": "3.12.3",
  "date_time": {...}
}
```

## Operations

- `facts.all.get` (query)
- `facts.category.get` (query)

## Provider

- `internal/provider/node/facts/`
- Aggregates data from existing providers (host, disk, mem, load) plus
  additional hardware detection (CPU model, virtualization type)
- Use `lscpu`, `dmidecode`, `systemd-detect-virt`

## Notes

- This is essentially what Ansible gathers on every connection
- Cache results (facts don't change frequently) with TTL
- No auth for basic facts; detailed hardware may need `facts:read`
- Useful for fleet management when querying multiple hosts via `_all`
