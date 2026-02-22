---
title: Document feature roadmap with tiered priorities
status: done
created: 2026-02-15
updated: 2026-02-18
---

## Objective

Add a Docusaurus page documenting the feature roadmap organized by priority
tiers. This gives users and contributors a clear picture of where OSAPI is
headed and what's needed to be a legitimate OS management appliance.

## Location

`docs/docs/sidebar/roadmap.md` — add to sidebar navigation.

## Content Structure

### Current Capabilities (what exists today)

- System: hostname, status, uptime, OS info, disk, memory, load
- Network: DNS get/update, ping
- Job system: async job processing via NATS JetStream
- Auth: JWT bearer tokens with role-based scopes
- CLI: full parity with API

### Tier 1 — Core Appliance

The minimum to be taken seriously as an OS management API.

| Feature               | Description                                 | Ansible Equivalent      |
| --------------------- | ------------------------------------------- | ----------------------- |
| Service management    | systemctl start/stop/restart/enable/disable | `service`, `systemd`    |
| Package management    | Install, remove, update packages            | `apt`, `yum`, `package` |
| User/group management | Create, modify, delete users and groups     | `user`, `group`         |
| Power management      | Shutdown, reboot (with delay/scheduling)    | `reboot`                |
| Hostname set          | Set hostname (complement existing get)      | `hostname`              |
| Health endpoints      | Liveness, readiness, detailed health        | —                       |

### Tier 2 — Security & Networking

What makes it production-ready and secure.

| Feature             | Description                              | Ansible Equivalent    |
| ------------------- | ---------------------------------------- | --------------------- |
| Firewall management | ufw/nftables rule management             | `ufw`, `firewalld`    |
| Network interfaces  | IP config, routing, interface up/down    | `nmcli`               |
| SSH key management  | Authorized key management per user       | `authorized_key`      |
| TLS certificates    | Certificate install, CSR, CA trust store | `openssl_certificate` |
| SELinux/AppArmor    | Security policy mode and profiles        | `selinux`             |
| Audit logging       | Structured API operation audit trail     | —                     |

### Tier 3 — Operations & Observability

What makes it useful for day-to-day operations.

| Feature              | Description                          | Ansible Equivalent           |
| -------------------- | ------------------------------------ | ---------------------------- |
| File management      | Read, write, lineinfile, permissions | `file`, `copy`, `lineinfile` |
| Command execution    | Ad-hoc command/shell execution       | `command`, `shell`           |
| Process management   | List, inspect, signal processes      | —                            |
| Log viewing          | Query systemd journal and syslog     | —                            |
| NTP/time management  | NTP sync, timezone configuration     | `chrony`, `timezone`         |
| System updates       | Check and apply OS patches           | `apt upgrade`                |
| Sysctl/kernel params | Query and tune kernel parameters     | `sysctl`                     |

### Tier 4 — Advanced

Differentiators for fleet management and enterprise use.

| Feature                | Description                             | Ansible Equivalent |
| ---------------------- | --------------------------------------- | ------------------ |
| System facts/inventory | Comprehensive hardware/OS/network facts | `setup`            |
| Storage management     | LVM, mounts, SMART health               | `lvol`, `mount`    |
| Cron/scheduling        | Scheduled task management               | `cron`             |

### Implementation Pattern

Each feature follows the same architecture:

1. Provider interface + platform implementations
2. Job operation types and subject routing
3. Worker processor dispatch
4. Job client methods
5. OpenAPI spec with strict-server + BearerAuth
6. API handler with scope middleware
7. CLI commands with `--json` output
8. Tests (provider, client, handler, integration)

### Contributing

Link to contributing guide and explain how to pick up a feature from the
roadmap.

## Notes

- Keep the page updated as features are implemented (move from roadmap to
  "Current Capabilities")
- Consider adding status badges (planned, in progress, complete)
- Reference individual `.tasks/backlog/` files for detailed specs
- This page should be the canonical "what can OSAPI do and where is it going"
  resource

## Outcome

Created `docs/docs/sidebar/roadmap.md` at sidebar position 9 with all four
priority tiers, current capabilities summary, implementation pattern, and
contributing section.
