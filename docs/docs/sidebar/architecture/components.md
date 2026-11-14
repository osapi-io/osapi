---
sidebar_position: 2
---

# Components

osapi is one binary that runs three processes. They can share a host or be
spread across many. `osapi start` runs all three in one process, which is how a
single host is usually set up; `osapi nats server start`,
`osapi controller start` and `osapi agent start` run them apart.

## NATS server

The NATS server is where everything else meets. It carries every job from the
controller to the agents and every result back, and it holds osapi's state in
JetStream: the jobs, their results, the agent registry, uploaded files and the
audit log. [State and storage](state.md) lists each one.

osapi embeds a NATS server with JetStream, so nothing else needs installing. It
listens on port 4222 by default (`nats.server.port`) and keeps JetStream's files
in `nats.server.store_dir`. A fleet larger than one site, or one that needs NATS
to survive a host, uses an external NATS cluster instead: point
`controller.nats` and `agent.nats` at it and do not start the embedded one.

The server authenticates its clients with no auth, a user and password, or an
NKey (`nats.server.auth`), and each client says which in its own `nats.auth`
section.

## Controller

The controller is the control plane. It never runs a command on a host and needs
no privileges on one. It runs:

- **The REST API** on port 8080 (`controller.api.port`). It checks the token on
  every request, validates the input, turns the request into a job and waits up
  to `controller.api.job_timeout` (30 seconds) for the result. Routes are under
  `/api/`: `/api/node/{hostname}/...` for work on hosts, and `/api/job`,
  `/api/agent`, `/api/audit`, `/api/file` and `/api/health` for the controller's
  own resources. [How a request runs](request-flow.md) follows one.
- **The management UI**, embedded in the binary and served by the same HTTP
  server when `controller.ui.enabled` is set. See [UI](ui.md).
- **The audit log**: every authenticated request is recorded to its own stream.
  See [Security](security.md#audit).
- **Enrollment**, when PKI is on: it holds the agents waiting to be accepted and
  signs every job it queues. See [Security](security.md#which-agent-may-answer).
- **A heartbeat** into the registry, so `osapi client health status` can show
  the controller and each of its parts, and **a condition watcher** that turns
  changes in any component's conditions into notifications. See
  [Observability](observability.md).
- **A metrics server** on its own port, 9090 by default
  (`controller.metrics.port`), with `/metrics`, `/health` and `/health/ready`.

## Agent

The agent runs on each managed host and does the work. At startup it:

- reads its **machine ID** from `/etc/machine-id` (or the platform UUID on
  macOS), and refuses to start without one. The machine ID is the agent's
  identity; the hostname is only its display name and may change;
- when PKI is on, generates or loads its keypair and **enrolls** with the
  controller, waiting until an administrator accepts it;
- when privilege escalation is on, runs a **preflight** that checks it may run
  each privileged program with sudo and holds the capabilities it needs, and
  exits if not;
- builds the **provider** for each behavior and domain on its platform, and logs
  which behaviors are unsupported here;
- subscribes to the job subjects for its machine ID, `_any`, `_all` and each of
  its labels.

While it runs, it writes a heartbeat to the registry every 10 seconds, collects
**facts** about the host (interfaces, routes, memory and the like) every 60
seconds (`agent.facts.interval`), evaluates its conditions such as disk and
memory pressure, and takes up to `agent.max_jobs` jobs at once.

The agent runs as an unprivileged user by default. Operations that change the
host then fail with a permission error rather than doing anything unexpected. To
let it change the host, either run it as root or turn on
`agent.privilege_escalation`, which runs only the writes through sudo with a
fixed sudoers drop-in; [agent hardening](../features/agent-hardening.md) has the
entries.

## Which process needs what

| Process     | Listens on                            | Needs on the host                                 | Holds state in                 |
| ----------- | ------------------------------------- | ------------------------------------------------- | ------------------------------ |
| NATS server | 4222, and a metrics port              | A directory for JetStream                         | JetStream                      |
| Controller  | 8080 for the API and UI, 9090 metrics | Nothing privileged                                | NATS                           |
| Agent       | A metrics port, when enabled          | Root, or sudo for the writes it is configured for | NATS, and the files it manages |

## Built on

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

## Design details

- [The job system](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md)
  and
  [transport](https://github.com/osapi-io/specs/blob/main/components/osapi/transport.md),
  for how the processes talk.
- [Providers](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md)
  and
  [running commands](https://github.com/osapi-io/specs/blob/main/components/osapi/exec.md),
  for what the agent does on a host.

[Cobra]: https://github.com/spf13/cobra
[Echo]: https://echo.labstack.com
[Viper]: https://github.com/spf13/viper
[NATS]: https://nats.io
[oapi-codegen]: https://github.com/oapi-codegen/oapi-codegen
[gopsutil]: https://github.com/shirou/gopsutil
[pro-bing]: https://github.com/prometheus-community/pro-bing
[golang-jwt]: https://github.com/golang-jwt/jwt
[OpenTelemetry]: https://opentelemetry.io
