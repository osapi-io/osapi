---
sidebar_position: 2
---

# System Architecture

OSAPI is a Linux system management platform that exposes a REST API for querying
and modifying host configuration and uses NATS JetStream for distributed,
asynchronous job processing. Operators interact with the system through a CLI
that can either hit the REST API directly or manage the job queue.

This page covers what an operator needs in order to run and call OSAPI: the
health check endpoints, how requests are authenticated and authorized, and what
the system depends on. How the code is laid out internally is a contributor's
question and is answered in the specifications repository — see Further Reading.

## Health Checks (`internal/controller/api/health/`)

The controller exposes three health check endpoints following the Kubernetes
liveness/readiness probe pattern. Liveness and readiness probes are
unauthenticated and live outside the authenticated API surface because they
serve infrastructure concerns rather than business operations. The detailed
system status endpoint requires JWT authentication with the `health:read`
permission. See the [API reference](/category/api) for exact paths and response
schemas.

### Liveness

Returns `{"status":"ok"}` unconditionally. No dependency checks are performed.
If the HTTP server responds, the process is alive. This endpoint is deliberately
trivial — putting dependency checks here would cause orchestrators to restart
the process during a transient NATS outage, creating a restart storm on top of
the original problem.

### Readiness

Runs all checks registered with the `Checker` interface and returns 200
(`ready`) or 503 (`not_ready`). The default checker (`NATSChecker`) verifies:

- **NATS connectivity** — the NATS connection is active and has a connected URL
- **KV bucket access** — the `job-queue` KV bucket is reachable and can list
  keys

Load balancers should use this endpoint to decide whether to route traffic. When
readiness fails, the server stays running but stops receiving requests until the
dependency recovers.

### Status

Breaks out each dependency as a named component with its own status and error
message. Also reports NATS connection info, JetStream stream statistics, KV
bucket statistics, job queue counts, application version, and uptime. Returns
`ok` when all components are healthy or `degraded` (with HTTP 503) when any
component fails. Requires JWT authentication because it exposes internal
topology.

Components checked:

| Component | What it checks                      |
| --------- | ----------------------------------- |
| `nats`    | NATS client is connected            |
| `kv`      | `job-queue` KV bucket is accessible |

Additional metrics (optional, gracefully skipped on failure):

| Section   | What it reports                                        |
| --------- | ------------------------------------------------------ |
| `nats`    | Connected URL, server version                          |
| `streams` | Message count, bytes, consumer count                   |
| `kv`      | Bucket name, key count, bytes                          |
| `jobs`    | Total, unprocessed, processing, completed, failed, DLQ |

### CLI Access

Operators can check health from the command line:

```bash
osapi client health              # liveness
osapi client health ready        # readiness
osapi client health status       # system status with metrics (requires auth)
```

## Security

### Authentication

The API uses **JWT HS256** tokens signed with a shared secret
(`security.signing_key`). Tokens carry a `roles` claim (array) that determines
the caller's access level. The `osapi token generate` command creates tokens for
a given role. Tokens can also carry a `permissions` claim that overrides
role-based expansion.

### Authorization

Access control uses fine-grained `resource:verb` permissions. Each API endpoint
declares a required permission (e.g., `node:read`, `schedule:write`,
`command:execute`). Built-in roles (`admin`, `write`, `read`) expand to default
permission sets, and custom roles can be defined in config. See
[Authentication & RBAC](../features/authentication.md) for the full permission
model.

The health endpoints `/health` and `/health/ready` are exceptions — they bypass
JWT authentication so that load balancers and orchestrators can probe them
without credentials.

### CORS

Cross-Origin Resource Sharing is configured per-server via
`controller.api.security.cors.allow_origins` in `osapi.yaml`. An empty list
disables CORS headers entirely.

## External Dependencies

| Dependency                    | Purpose                                      |
| ----------------------------- | -------------------------------------------- |
| [Echo][]                      | HTTP framework for the REST API              |
| [Cobra][] / [Viper][]         | CLI framework and configuration              |
| [NATS][] / JetStream          | Messaging, KV store, stream processing       |
| [oapi-codegen][]              | OpenAPI strict-server code generation        |
| [OpenTelemetry][]             | Distributed tracing and Prometheus metrics   |
| [gopsutil][]                  | Cross-platform system metrics                |
| [pro-bing][]                  | ICMP ping implementation                     |
| [golang-jwt][]                | JWT creation and validation                  |
| `nats-client` / `nats-server` | Sibling repos, pinned by version in `go.mod` |

## Further Reading

- [Running Jobs](job-architecture.md) — targeting, statuses, polling, and what
  to monitor
- [Adding an API Domain](../development/adding-an-api-domain.md) — the layers,
  the API design guidelines, and the design principles, stated in the
  specifications repository and indexed there
- [Contributing](https://github.com/osapi-io/osapi/blob/main/CONTRIBUTING.md) —
  setup, building, testing, and the conventions code follows

<!-- prettier-ignore-start -->
[Cobra]: https://github.com/spf13/cobra
[Echo]: https://echo.labstack.com
[Viper]: https://github.com/spf13/viper
[NATS]: https://nats.io
[oapi-codegen]: https://github.com/oapi-codegen/oapi-codegen
[gopsutil]: https://github.com/shirou/gopsutil
[pro-bing]: https://github.com/prometheus-community/pro-bing
[golang-jwt]: https://github.com/golang-jwt/jwt
[OpenTelemetry]: https://opentelemetry.io
<!-- prettier-ignore-end -->
