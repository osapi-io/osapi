---
sidebar_position: 3
---

# Features

OSAPI provides a comprehensive set of features for managing Linux systems.

<!-- prettier-ignore-start -->

|     | Feature                                        | Description                                                                                   |
| --- | ---------------------------------------------- | --------------------------------------------------------------------------------------------- |
| 🖥️  | [Node Management](node-management.md)          | Hostname, uptime, OS info, disk, memory, load                                                 |
| 🌐  | [Network Management](network-management.md)    | DNS read/update, ping                                                                         |
| ⚙️  | [Command Execution](command-execution.md)      | Remote exec and shell across managed hosts                                                    |
| 📁  | [File Management](file-management.md)          | Upload, deploy, and template files with SHA-based idempotency                                 |
| 📊  | [System Facts](system-facts.md)                | Agent-collected system facts -- architecture, kernel, FQDN, CPUs, network interfaces          |
| 🔄  | [Agent Lifecycle](agent-lifecycle.md)          | Node conditions, graceful drain/cordon for maintenance                                        |
| ⚡  | [Job System](job-system.md)                    | NATS JetStream with KV-first architecture -- broadcast, load-balanced, and label-based routing |
| 💚  | [Health Checks](health-checks.md)              | Liveness, readiness, system status endpoints                                                  |
| 📈  | [Metrics](metrics.md)                          | Prometheus `/metrics` endpoint                                                                |
| 📋  | [Audit Logging](audit-logging.md)              | Structured API audit trail with 30-day retention                                              |
| 🔐  | [Authentication & RBAC](authentication.md)     | JWT with fine-grained `resource:verb` permissions                                             |
| 📦  | [Container Management](container-management.md) | Docker lifecycle, exec, and pull through pluggable runtime drivers                             |
| ⏰  | [Cron Management](cron-management.md)          | Cron drop-in file and periodic script management                                              |
| 🔧  | [Sysctl Management](sysctl-management.md)      | Kernel parameter management via `/etc/sysctl.d/`                                              |
| 🕐  | [NTP Management](ntp-management.md)            | Chrony NTP server configuration and sync status                                               |
| 🌍  | [Timezone Management](timezone-management.md)  | System timezone get and set via timedatectl                                                   |
| 🔔  | [Notifications](notifications.md)              | Pluggable condition alerts with re-notification                                               |
| 🔍  | [Distributed Tracing](distributed-tracing.md)  | OpenTelemetry with trace context propagation                                                  |
| ⚡  | [Power Management](power-management.md)        | Reboot and shutdown target hosts with optional delay                                          |

<!-- prettier-ignore-end -->
