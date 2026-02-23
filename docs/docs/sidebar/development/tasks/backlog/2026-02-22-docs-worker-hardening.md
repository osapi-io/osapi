---
title: Document worker hardening and least-privilege deployment
status: backlog
created: 2026-02-22
updated: 2026-02-22
---

## Objective

Write a deployment guide for running the OSAPI worker with minimal privileges
using systemd sandboxing and AppArmor. The guide should cover capability
management, filesystem restrictions, command whitelisting, and resource limits.

This is documentation, not code. The worker already enforces `command:execute`
permissions via RBAC. OS-level hardening is a deployment concern that layers
defense in depth on top.

## Content

### Least-Privilege Capabilities

The worker needs specific Linux capabilities for certain operations. Document
how to grant only what's needed via systemd instead of running as root or using
`setcap` on the binary:

- **Ping** requires `CAP_NET_RAW`. Two approaches:
  - `AmbientCapabilities=CAP_NET_RAW` in the systemd unit (preferred, survives
    binary updates)
  - `sudo setcap cap_net_raw=+ep ./osapi` on the binary (must be reapplied after
    every build/deploy)
  - `sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"` as a system-wide
    alternative (allows all users to ping)
- Document which capabilities each OSAPI operation requires
- Show how to run the worker as a dedicated non-root user with only the
  capabilities it needs

### systemd Unit Hardening

Document recommended systemd directives for the worker service:

- `User=osapi` / `Group=osapi` — dedicated service account
- `AmbientCapabilities=CAP_NET_RAW` — grant only needed capabilities
- `CapabilityBoundingSet=CAP_NET_RAW` — drop everything else
- `NoNewPrivileges=yes` — child processes (executed commands) cannot gain
  additional privileges
- `ProtectSystem=strict` — read-only filesystem except allowed paths
- `ReadWritePaths=` — whitelist paths the worker needs to write (e.g., NATS
  store dir, temp dirs)
- `ProtectHome=yes` — no access to /home
- `InaccessiblePaths=` — block sensitive paths entirely
- `PrivateTmp=yes` — isolated /tmp for the worker and its children
- `SystemCallFilter=` — restrict available syscalls
- `MemoryMax=` / `CPUQuota=` — resource limits for the worker and executed
  commands
- `ExecPaths=` / `NoExecPaths=` — whitelist which binaries the worker can
  execute (systemd 254+, command whitelisting at the OS level)

Provide a complete example unit file with sensible defaults and comments
explaining each directive.

### AppArmor Profile

Document an AppArmor profile for the worker that restricts:

- Which binaries the worker can `exec()` (command whitelisting)
- Which paths the worker can read/write
- Network access scope
- Capability restrictions

Provide a sample profile and instructions for loading/enforcing it.

### Placement

- `docs/docs/sidebar/deployment/worker-hardening.md` or similar
- Link from the command execution feature page
- Link from the configuration reference

## Notes

- Target audience is operators deploying OSAPI in production
- Should cover Debian/Ubuntu (AppArmor) as primary; mention SELinux for RHEL as
  an alternative
- Keep examples copy-pasteable with comments explaining each directive
- Document the capability requirements for each OSAPI provider so operators know
  exactly what to grant
