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
them to be used as appliances. You install a single binary, point it at a config
file, and get a REST API and CLI for querying and changing system
configuration — hostname, DNS, disk usage, memory, load averages, and more.
State-changing operations run asynchronously through a job queue so the API
server itself never needs root privileges.

<br clear="left"/>

## Features

- **System management** — query hostname, uptime, OS info, disk usage, memory,
  and load averages via REST API or CLI
- **Network management** — read and update DNS configuration, ping remote hosts
- **Async job system** — state-changing operations run through NATS JetStream
  with KV-first architecture, supporting broadcast, load-balanced, and
  label-based routing across multi-host deployments
- **Health checks** — liveness, readiness, and detailed system status endpoints
  for load balancers and monitoring
- **Audit logging** — structured audit trail of all API operations stored in
  NATS KV with 30-day retention, admin-only read access
- **Fine-grained RBAC** — `resource:verb` permissions with built-in roles
  (admin/write/read), custom role definitions, and direct permission grants
- **Prometheus metrics** — standard `/metrics` endpoint for monitoring
- **Distributed tracing** — OpenTelemetry integration with trace context
  propagation across HTTP and NATS boundaries
- **CLI parity** — every API operation has an equivalent CLI command with
  `--json` output for scripting
- **Namespace isolation** — multiple OSAPI deployments can share a single NATS
  cluster via subject and infrastructure prefixing

## Documentation

[Architecture][] | [Getting Started][] | [API][] | [Usage][] | [Roadmap][]

[Architecture]: https://osapi-io.github.io/osapi/sidebar/architecture
[Getting Started]: https://osapi-io.github.io/osapi/
[API]: https://osapi-io.github.io/osapi/category/api
[Usage]: https://osapi-io.github.io/osapi/sidebar/usage/
[Roadmap]: https://osapi-io.github.io/osapi/sidebar/roadmap

## License

The [MIT][] License.

[MIT]: LICENSE
