---
sidebar_position: 10
---

# Roadmap

OSAPI aims to be a comprehensive Linux system management API. This page
documents what exists today and where the project is headed, organized by
priority tier.

## Current Capabilities

These features are implemented and available today:

- **System** — hostname, status, uptime, OS info, disk, memory, load
- **Network** — DNS get/update, ping
- **Job system** — async job processing via NATS JetStream with KV-first
  architecture, broadcast/load-balanced/label-based routing
- **Authentication** — JWT bearer tokens with role-based scopes
  (admin/write/read)
- **CLI** — full parity with the REST API

## Tier 1 — Core Appliance

The minimum feature set to be taken seriously as an OS management API.

| Feature               | Description                                 | Ansible Equivalent      |
| --------------------- | ------------------------------------------- | ----------------------- |
| Service management    | systemctl start/stop/restart/enable/disable | `service`, `systemd`    |
| Package management    | Install, remove, update packages            | `apt`, `yum`, `package` |
| User/group management | Create, modify, delete users and groups     | `user`, `group`         |
| Power management      | Shutdown, reboot (with delay/scheduling)    | `reboot`                |
| Hostname set          | Set hostname (complement existing get)      | `hostname`              |
| Health endpoints      | Liveness, readiness, system status          | —                       |

## Tier 2 — Security & Networking

What makes OSAPI production-ready and secure.

| Feature             | Description                              | Ansible Equivalent    |
| ------------------- | ---------------------------------------- | --------------------- |
| Firewall management | ufw/nftables rule management             | `ufw`, `firewalld`    |
| Network interfaces  | IP config, routing, interface up/down    | `nmcli`               |
| SSH key management  | Authorized key management per user       | `authorized_key`      |
| TLS certificates    | Certificate install, CSR, CA trust store | `openssl_certificate` |
| SELinux/AppArmor    | Security policy mode and profiles        | `selinux`             |
| Audit logging       | Structured API operation audit trail     | —                     |

## Tier 3 — Operations & Observability

What makes OSAPI useful for day-to-day operations.

| Feature              | Description                          | Ansible Equivalent           |
| -------------------- | ------------------------------------ | ---------------------------- |
| File management      | Read, write, lineinfile, permissions | `file`, `copy`, `lineinfile` |
| Command execution    | Ad-hoc command/shell execution       | `command`, `shell`           |
| Process management   | List, inspect, signal processes      | —                            |
| Log viewing          | Query systemd journal and syslog     | —                            |
| NTP/time management  | NTP sync, timezone configuration     | `chrony`, `timezone`         |
| System updates       | Check and apply OS patches           | `apt upgrade`                |
| Sysctl/kernel params | Query and tune kernel parameters     | `sysctl`                     |

## Tier 4 — Advanced

Differentiators for fleet management and enterprise use.

| Feature                | Description                             | Ansible Equivalent |
| ---------------------- | --------------------------------------- | ------------------ |
| System facts/inventory | Comprehensive hardware/OS/network facts | `setup`            |
| Storage management     | LVM, mounts, SMART health               | `lvol`, `mount`    |
| Cron/scheduling        | Scheduled task management               | `cron`             |

## Implementation Pattern

Each new feature follows the same architecture:

1. Provider interface + platform implementations
2. Job operation types and subject routing
3. Worker processor dispatch
4. Job client methods
5. OpenAPI spec with strict-server + BearerAuth
6. API handler with scope middleware
7. CLI commands with `--json` output
8. Tests (provider, client, handler, integration)

See [Job System Architecture](architecture/job-architecture.md) for details on
the provider and worker pipeline.

## Contributing

Want to pick up a feature from the roadmap? Start with the
[Contributing](contributing.md) guide, then:

1. Open an issue or discussion to claim the feature
2. Follow the implementation pattern above
3. Submit a PR with tests and documentation

Lower-tier features are higher priority, but contributions at any tier are
welcome.
