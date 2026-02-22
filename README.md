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

## Features

| | |
|---|---|
| 🖥️ **System & Network** | Hostname, uptime, OS info, disk, memory, load, DNS read/update, ping, command execution (exec/shell) |
| ⚡ **Async Job System** | NATS JetStream with KV-first architecture — broadcast, load-balanced, and label-based routing across hosts |
| 💚 **Health & Metrics** | Liveness, readiness, system status endpoints, Prometheus `/metrics` |
| 📋 **Audit Logging** | Structured API audit trail in NATS KV with 30-day retention and admin-only read access |
| 🔐 **Auth & RBAC** | JWT with fine-grained `resource:verb` permissions, built-in and custom roles, direct permission grants |
| 🔍 **Distributed Tracing** | OpenTelemetry with trace context propagation across HTTP and NATS |
| 🖥️ **CLI Parity** | Every API operation has a CLI equivalent with `--json` for scripting |
| 🏢 **Multi-Tenant** | Namespace isolation lets multiple deployments share a single NATS cluster |

## Documentation

[Features][] | [Architecture][] | [Getting Started][] | [API][] | [Usage][] | [Roadmap][]

[Features]: https://osapi-io.github.io/osapi/category/features
[Architecture]: https://osapi-io.github.io/osapi/sidebar/architecture
[Getting Started]: https://osapi-io.github.io/osapi/
[API]: https://osapi-io.github.io/osapi/category/api
[Usage]: https://osapi-io.github.io/osapi/sidebar/usage
[Roadmap]: https://osapi-io.github.io/osapi/sidebar/development/roadmap

## License

The [MIT][] License.

[MIT]: LICENSE
