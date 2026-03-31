---
sidebar_position: 1
sidebar_label: Overview
---

# Architecture

OSAPI turns Linux servers into managed appliances. You install a single binary,
point it at a config file, and get a REST API and CLI for querying and changing
system configuration — hostname, DNS, disk usage, memory, load averages, and
more. State-changing operations run asynchronously through a job queue so the
controller itself never needs root privileges.

## The Three Processes

OSAPI has three runtime components. They can all run on the same host or be
spread across many.

### NATS Server

A lightweight message broker that stores job state and routes messages between
the controller and agents. OSAPI embeds a NATS server with JetStream enabled, so
you don't need to install anything extra — just run `osapi nats server start`.

For production deployments with multiple hosts, you can point everything at an
external NATS cluster instead of the embedded one. Just change the `nats.server`
section in `osapi.yaml`.

### Controller

The control plane process. It runs several sub-components:

- **REST API** — an HTTP server that handles authentication (JWT), validates
  requests, and translates them into jobs published to NATS. The controller
  never executes system commands directly — it creates a job and returns a job
  ID. Clients poll for results.
- **Component heartbeat** — registers the controller in the registry KV so
  `health status` can report its state. The heartbeat includes sub-component
  status (api, metrics, notifier, tracing) so operators can see the health of
  each internal service.
- **Condition watcher** — monitors the registry KV for condition transitions and
  dispatches notifications.

Start it with `osapi controller start`.

### Agent

A background process that subscribes to NATS, picks up jobs, and executes the
actual system operations (reading hostname, querying DNS, checking disk usage,
etc.). Agents run with whatever privileges they have — if an agent can't read
something due to permissions, it reports the error rather than failing silently.
Each agent publishes its own sub-component status (heartbeat, metrics) via the
registry heartbeat.

Start it with `osapi agent start`.

## Deployment Models

### Single Host

The simplest setup. All three processes run on the same machine:

```mermaid
graph TD
    subgraph host["Linux Host"]
        CLI["CLI"]
        API["Controller"]
        Agent["Agent"]
        NATS["NATS (embedded)"]

        CLI -->|HTTP| API
        API -->|publish job| NATS
        NATS -->|deliver job| Agent
        Agent -->|write result| NATS
        API -->|read result| NATS
    end
```

Use `osapi start` to run all three in a single process — the recommended
approach for single-host deployments:

```bash
osapi start
```

The CLI on the same host talks to the controller over localhost. This is useful
for managing a single appliance or for development.

### Multi-Host

For managing a fleet, run a shared NATS server (or cluster) and point multiple
agents at it. Each agent registers with its hostname and optional labels, and
the job routing system delivers work to the right place.

```mermaid
graph TD
    CLI["CLI"]
    API["Controller"]
    NATS["NATS (shared)"]
    W1["Agent (web-01)"]
    W2["Agent (web-02)"]

    CLI -->|HTTP| API
    API -->|publish job| NATS
    NATS -->|deliver job| W1
    NATS -->|deliver job| W2
    W1 -->|write result| NATS
    W2 -->|write result| NATS
    API -->|read result| NATS
```

You can target jobs to specific hosts, broadcast to all, or route by label:

- `--target _any` — send to any available agent (load balanced)
- `--target _all` — send to every agent (broadcast)
- `--target web-01` — send to a specific host
- `--target group:web.dev` — send to all agents with a matching label

## How a Request Flows

When you run a command like `osapi client node hostname`:

```mermaid
sequenceDiagram
    participant CLI
    participant API as Controller
    participant NATS
    participant Agent

    CLI->>API: GET /api/v1/node/{hostname}/hostname
    API->>NATS: store job in KV
    API->>NATS: publish notification
    NATS->>Agent: deliver notification
    Agent->>NATS: read job from KV
    Agent->>Agent: execute operation
    Agent->>NATS: write result to KV
    API->>NATS: read result from KV
    API-->>CLI: 200 (result + job_id)
```

The controller never touches the operating system directly. It's a thin
coordination layer between clients and agents.

## Further Reading

For details on individual features — what they do, how they work, and how to
configure them — see the Features section:

- [Node Management](../features/node-management.md) — hostname, disk, memory,
  load
- [Network Management](../features/network-management.md) — DNS, ping
- [Command Execution](../features/command-execution.md) — exec, shell
- [File Management](../features/file-management.md) — upload, deploy, templates
- [Container Management](../features/container-management.md) — Docker
  lifecycle, exec, pull
- [Cron Management](../features/cron-management.md) — cron drop-in file
  management
- [Sysctl Management](../features/sysctl-management.md) — kernel parameter
  management
- [NTP Management](../features/ntp-management.md) — NTP server management
- [Timezone Management](../features/timezone-management.md) — system timezone
- [Power Management](../features/power-management.md) — reboot and shutdown
- [Process Management](../features/process-management.md) — list, inspect, and
  signal processes
- [Job System](../features/job-system.md) — async job processing and routing
- [Audit Logging](../features/audit-logging.md) — API audit trail and export
- [Health Checks](../features/health-checks.md) — liveness, readiness, status
- [Authentication & RBAC](../features/authentication.md) — JWT, roles,
  permissions
- [Distributed Tracing](../features/distributed-tracing.md) — OpenTelemetry
  integration
- [Metrics](../features/metrics.md) — Prometheus endpoint

## Deep Dives

- [System Architecture](system-architecture.md) — package layout, handler
  structure, provider pattern, and code-level details
- [Job Architecture](job-architecture.md) — KV-first design, subject routing,
  agent pipeline, and multi-host processing
- [Configuration](../usage/configuration.md) — full `osapi.yaml` reference with
  every supported field
- [API Design Guidelines](api-guidelines.md) — REST conventions and endpoint
  patterns
- [Guiding Principles](principles.md) — design philosophy and project values
- [Development](../development/development.md) — setup, building, testing, and
  contributing
