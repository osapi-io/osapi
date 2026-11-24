---
sidebar_position: 2
---

# Components

osapi is one binary that runs three processes. The processes can be on one host
or on many hosts. `osapi start` runs the three processes in one process. Use it
on a single host. `osapi nats start`, `osapi controller start` and
`osapi agent start` run each process alone.

## NATS server

The NATS server connects all the other parts. It carries each job from the
controller to the agents, and each result back. It also keeps the state of osapi
in JetStream: the jobs, their results, the agent registry, the uploaded files
and the audit log. [State and storage](state.md) lists each one.

osapi has an embedded NATS server with JetStream, so you do not need to install
NATS. The server listens on port 4222 by default (`nats.server.port`). It keeps
the JetStream files in `nats.server.store_dir`. With
`nats.server.jetstream.key_file`, it encrypts them.
[Encryption at rest](../usage/configuration.md#encryption-at-rest) tells how.

Use an external NATS cluster if the hosts are in more than one site, or if NATS
must continue to operate when a host stops. Set `controller.nats` and
`agent.nats` to the address of the cluster. Do not start the embedded server.

The server authenticates its clients in one of these modes (`nats.server.auth`):

- no authentication
- a user name and a password
- an NKey
- operator mode, with a credentials file for each user

Each client sets its own mode in its `nats.auth` section.

## Controller

The controller is the control plane. It does not run a command on a host, and it
needs no privileges on a host. It runs these parts:

- **The REST API** on port 8080 (`controller.api.port`). The API checks the
  token of each request and validates the input. Then it makes a job from the
  request and waits for the result. It waits up to `controller.api.job_timeout`
  (30 seconds). The routes are under `/api/`. `/api/node/{hostname}/...` is for
  work on hosts. `/api/job`, `/api/agent`, `/api/audit`, `/api/file` and
  `/api/health` are for the resources of the controller.
  [How a request runs](request-flow.md) follows one request.
- **The management UI.** It is in the binary, and the same HTTP server sends it
  when `controller.ui.enabled` is set. Refer to [UI](ui.md).
- **The audit log.** The controller records each authenticated request in its
  own stream. Refer to [Security](security.md#audit).
- **Enrollment**, when PKI is on. The controller keeps the agents that wait for
  acceptance, and it signs each job that it puts on the queue. Refer to
  [Security](security.md#which-agent-can-answer).
- **A heartbeat** in the registry. With it, `osapi client health status` shows
  the controller and each of its parts.
- **A condition watcher.** It sends a notification when the conditions of a
  component change. Refer to [Observability](observability.md).
- **A metrics server** on its own port, 9090 by default
  (`controller.metrics.port`), with `/metrics`, `/health` and `/health/ready`.

## Agent

The agent runs on each managed host and does the work. When it starts, it does
these steps:

1. It reads its **machine ID** from `/etc/machine-id`, or the platform UUID on
   macOS. If it has no machine ID, it does not start. The machine ID is the
   identity of the agent. The hostname is only a display name, and it can
   change.
1. If PKI is on, it makes or loads its key pair and **enrolls** with the
   controller. It waits until an administrator accepts it.
1. If privilege escalation is on, it runs a **preflight**. If sudo needs a
   password, the agent stops. The agent shows a warning for each installed
   program that has no sudo rule, and for each missing capability.
1. It makes the **provider** for each behavior and domain on its platform. It
   writes to the log the behaviors that its platform does not support.
1. It subscribes to the job subjects for its machine ID, for `_any`, for `_all`
   and for each of its labels.

While the agent runs, it does these tasks:

- It writes a heartbeat to the registry every 10 seconds.
- It collects **facts** about the host every 60 seconds
  (`agent.facts.interval`), for example the interfaces, the routes and the
  memory.
- It examines its conditions, for example disk pressure and memory pressure.
- It runs up to `agent.max_jobs` jobs at the same time.

By default, the agent runs as a user without privileges. An operation that
changes the host then fails with a permission error. The operation does not do
something that you do not expect. To let the agent change the host, run it as
root, or enable `agent.privilege_escalation`. Privilege escalation runs only the
writes through sudo, with a fixed sudoers drop-in.
[Agent hardening](../features/platform/agent-hardening.md) gives the entries.

## What each process needs

| Process     | Listens on                            | Needs on the host                               | Keeps state in                 |
| ----------- | ------------------------------------- | ----------------------------------------------- | ------------------------------ |
| NATS server | 4222, and a metrics port              | A directory for JetStream                       | JetStream                      |
| Controller  | 8080 for the API and UI, 9090 metrics | No privileges                                   | NATS                           |
| Agent       | A metrics port, when it is enabled    | Root, or sudo for the writes that you configure | NATS, and the files it manages |

## Built on

| Dependency                    | Purpose                                             |
| ----------------------------- | --------------------------------------------------- |
| [Echo][]                      | The HTTP framework of the REST API                  |
| [Cobra][] / [Viper][]         | The CLI framework and the configuration             |
| [NATS][] / JetStream          | Messages, the KV store and streams                  |
| [oapi-codegen][]              | Code generation for the OpenAPI strict server       |
| [OpenTelemetry][]             | Distributed traces and Prometheus metrics           |
| [gohai][]                     | Every read of the host: facts, metrics, processes   |
| [golang-jwt][]                | JWT creation and validation                         |
| `nats-client` / `nats-server` | Related repositories, pinned by version in `go.mod` |

## Design details

- [The job system](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md)
  and
  [transport](https://github.com/osapi-io/specs/blob/main/components/osapi/transport.md)
  tell how the processes communicate.
- [Providers](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md)
  and
  [running commands](https://github.com/osapi-io/specs/blob/main/components/osapi/exec.md)
  tell what the agent does on a host.

[Cobra]: https://github.com/spf13/cobra
[Echo]: https://echo.labstack.com
[Viper]: https://github.com/spf13/viper
[NATS]: https://nats.io
[oapi-codegen]: https://github.com/oapi-codegen/oapi-codegen
[gohai]: https://github.com/osapi-io/gohai
[golang-jwt]: https://github.com/golang-jwt/jwt
[OpenTelemetry]: https://opentelemetry.io
