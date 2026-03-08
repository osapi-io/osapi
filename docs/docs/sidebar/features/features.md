---
sidebar_position: 3
---

# Features

OSAPI provides a comprehensive set of features for managing Linux
systems.

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
| 🔍  | [Distributed Tracing](distributed-tracing.md)  | OpenTelemetry with trace context propagation                                                  |

<!-- prettier-ignore-end -->
