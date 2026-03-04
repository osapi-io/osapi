[![release](https://img.shields.io/github/release/retr0h/osapi.svg?style=for-the-badge)](https://github.com/retr0h/osapi/releases/latest)
[![codecov](https://img.shields.io/codecov/c/github/retr0h/osapi?token=NF0T86B1EP&style=for-the-badge)](https://codecov.io/gh/retr0h/osapi)
[![go report card](https://goreportcard.com/badge/github.com/retr0h/osapi?style=for-the-badge)](https://goreportcard.com/report/github.com/retr0h/osapi)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](LICENSE)
[![build](https://img.shields.io/github/actions/workflow/status/retr0h/osapi/go.yml?style=for-the-badge)](https://github.com/retr0h/osapi/actions/workflows/go.yml)
[![powered by](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
[![conventional commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
![openapi initiative](https://img.shields.io/badge/openapiinitiative-%23000000.svg?style=for-the-badge&logo=openapiinitiative&logoColor=white)
![Linux](https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black)
![gitHub commit activity](https://img.shields.io/github/commit-activity/m/retr0h/osapi?style=for-the-badge)

# OS API

<img src="asset/logo.png" align="left" />

*(OSAPI /ˈoʊsɑːpi/ - Oh-sah-pee)* A CRUD API for managing Linux systems.

This project provides basic management capabilities to Linux systems, enabling
them to be used as appliances.

<br clear="left"/>

<img src="asset/demo.gif" alt="OSAPI demo" />

## ✨ Features

| | |
|---|---|
| 🖥️ **[Node Management][]** | Hostname, uptime, OS info, disk, memory, load |
| 🌐 **[Network Management][]** | DNS read/update, ping |
| ⚙️ **[Command Execution][]** | Remote exec and shell across managed hosts |
| ⚡ **[Async Job System][]** | NATS JetStream with KV-first architecture — broadcast, load-balanced, and label-based routing across hosts |
| 💚 **[Health][] & [Metrics][]** | Liveness, readiness, system status endpoints, Prometheus `/metrics` |
| 📋 **[Audit Logging][]** | Structured API audit trail in NATS KV with 30-day retention and admin-only read access |
| 🔐 **[Auth & RBAC][]** | JWT with fine-grained `resource:verb` permissions, built-in and custom roles, direct permission grants |
| 🔍 **[Distributed Tracing][]** | OpenTelemetry with trace context propagation across HTTP and NATS |
| 🖥️ **CLI Parity** | Every API operation has a CLI equivalent with `--json` for scripting |
| 🏢 **Multi-Tenant** | Namespace isolation lets multiple deployments share a single NATS cluster |

[Node Management]: https://osapi-io.github.io/osapi/sidebar/features/node-management
[Network Management]: https://osapi-io.github.io/osapi/sidebar/features/network-management
[Command Execution]: https://osapi-io.github.io/osapi/sidebar/features/command-execution
[Async Job System]: https://osapi-io.github.io/osapi/sidebar/features/job-system
[Health]: https://osapi-io.github.io/osapi/sidebar/features/health-checks
[Metrics]: https://osapi-io.github.io/osapi/sidebar/features/metrics
[Audit Logging]: https://osapi-io.github.io/osapi/sidebar/features/audit-logging
[Auth & RBAC]: https://osapi-io.github.io/osapi/sidebar/features/authentication
[Distributed Tracing]: https://osapi-io.github.io/osapi/sidebar/features/distributed-tracing

## 📖 Documentation

[Features][] | [Architecture][] | [Getting Started][] | [API][] | [Usage][] | [Roadmap][]

[Features]: https://osapi-io.github.io/osapi/category/features
[Architecture]: https://osapi-io.github.io/osapi/sidebar/architecture
[Getting Started]: https://osapi-io.github.io/osapi/
[API]: https://osapi-io.github.io/osapi/category/api
[Usage]: https://osapi-io.github.io/osapi/sidebar/usage
[Roadmap]: https://osapi-io.github.io/osapi/sidebar/development/roadmap

## 🔗 Sister Projects

| Project | Description |
| --- | --- |
| [osapi-sdk][] | Go SDK for OSAPI — client library and orchestration primitives |
| [osapi-orchestrator][] | A Go package for orchestrating operations across OSAPI-managed hosts — typed operations, chaining, conditions, and result decoding built on top of the osapi-sdk engine |
| [nats-client][] | A Go package for connecting to and interacting with a NATS server |
| [nats-server][] | A Go package for running an embedded NATS server |

[osapi-sdk]: https://github.com/osapi-io/osapi-sdk
[osapi-orchestrator]: https://github.com/osapi-io/osapi-orchestrator
[nats-client]: https://github.com/osapi-io/nats-client
[nats-server]: https://github.com/osapi-io/nats-server

## 📄 License

The [MIT][] License.

[MIT]: LICENSE
