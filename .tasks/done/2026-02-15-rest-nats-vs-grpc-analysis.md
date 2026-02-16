---
title: Architecture analysis — REST+NATS vs gRPC
status: done
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Evaluate whether osapi should keep its current REST+NATS architecture or switch
to gRPC. The goal: a clean, easy-to-extend API for turning stock Linux servers
into appliances — configuring firmware, NTP, RAID, DNS, etc. — without being
Ansible.

## Recommendation: Keep REST+NATS

**Short answer**: Keep the architecture. REST+NATS is the right choice for this
problem. gRPC would solve fewer problems than it creates.

## Why REST is right for the external API

Users are ops engineers and automation scripts, not microservices. REST gives
them:

- **curl/httpie/wget** — debug and explore with tools already on every server
- **Browser-based tools** — Swagger UI, Postman, any HTTP client
- **OpenAPI codegen** — generate Go, Python, Rust, TypeScript clients from the
  same spec
- **Idempotent semantics** — PUT/DELETE map naturally to "ensure this state"
  which is exactly the appliance config model
- **No client library required** — any language with HTTP support works

gRPC would require every consumer to install protoc tooling, generate stubs,
and use a gRPC client library. For a tool whose audience is "drop onto a Linux
server and configure it," that's a significant barrier.

The OpenAPI-first approach is a strength, not a liability. The contract is
defined, server and client are generated, and documentation comes for free. With
gRPC the same would be done with proto files — it's not simpler, just different.

## Why NATS is right for the job system

The job system solves real problems that gRPC doesn't address:

1. **Async operations** — Firmware updates, RAID rebuilds, and package installs
   take minutes. REST returns 202 + job ID, client polls for status. With gRPC
   the same pattern would need to be built (or use streaming, which is more
   complex).

2. **Fan-out routing** — The `_any` / `_all` / `{hostname}` subject routing is
   elegant. One API call targets a single host, load-balances across available
   workers, or broadcasts to a fleet. gRPC has no equivalent — a routing layer
   would need to be built on top.

3. **Durability** — JetStream gives persistent queues, replay, and exactly-once
   delivery. If a worker crashes mid-firmware-update, the job survives in KV.
   gRPC streams are ephemeral — a queue (like... NATS) would need to be bolted
   on for durability.

4. **Audit trail** — Append-only status events
   (`status.{job-id}.{event}.{hostname}.{nano}`) give a complete history of
   every operation. This matters for appliance management (compliance, debugging
   "who changed DNS at 3am").

5. **Decoupled privilege model** — The API server runs unprivileged; the worker
   runs with the privileges needed for system changes. NATS is the boundary.
   With gRPC, a different privilege separation mechanism would be needed.

## What gRPC would give

The honest benefits of gRPC for this project:

- **Streaming** — Server-push for real-time job status instead of polling. Nice
  but not essential. SSE on the REST API achieves the same with less complexity.
- **Strongly typed contracts** — Already have this via OpenAPI + codegen. Proto
  files are not meaningfully stronger than OpenAPI 3.x for this use case.
- **Performance** — Binary protobuf encoding is faster than JSON. But the
  bottleneck is "run a shell command to configure RAID" not "serialize a
  200-byte JSON payload."
- **Bidirectional streaming** — Useful for log tailing or real-time monitoring.
  But NATS subjects already give pub/sub for this.

None of these justify a rewrite.

## What gRPC would cost

- Rewrite the API layer — all handlers, middleware, auth, error mapping
- Rewrite the CLI client — from HTTP calls to gRPC stubs
- Lose curl-ability — the primary debugging and exploration tool disappears
  (grpcurl exists but is clunky)
- Lose OpenAPI ecosystem — documentation, client generation for non-Go
  languages, Swagger UI
- Still need async — gRPC doesn't solve the "firmware update takes 10 minutes"
  problem. NATS would still be needed, or a job queue in gRPC streaming would
  need to be built (reinventing what already exists, worse)
- Protoc toolchain — additional build dependency for code generation

## Where to focus instead

Rather than a rewrite, invest in:

1. **Complete the job migration (Phase 3-6)** — The current task-to-job
   migration is the right trajectory. The job system with KV-first storage,
   append-only events, and subject routing is well-designed. Finish it.

2. **Add SSE or WebSocket for real-time status** — If polling feels clunky for
   long-running jobs, add a streaming endpoint. This gives gRPC's main UX
   advantage without abandoning REST.

3. **Expand the provider surface** — The provider pattern
   (`internal/provider/`) is the extensibility model. Adding new system
   capabilities (NTP, RAID, firmware, packages) is purely additive — new
   provider implementations, new OpenAPI paths, new job operation types. The
   architecture supports this cleanly.

4. **Consider gRPC only for inter-service communication** — If multiple Go
   services on the same host later need to talk to each other (e.g., a
   monitoring agent talking to the config agent), gRPC makes sense there. But
   keep REST as the external API.

## Comparison

| Concern | REST+NATS | gRPC |
|---------|-----------|------|
| Ops-friendly debugging | curl, browser, Swagger | grpcurl (clunky) |
| Client generation | OpenAPI -> any language | Proto -> any language |
| Async job processing | NATS JetStream (built) | Must build or bolt on |
| Fan-out/routing | Subject hierarchy (built) | Must build |
| Durability | KV + streams (built) | Must build |
| Real-time push | Add SSE (incremental) | Streaming (native) |
| Privilege separation | API/Worker via NATS | Must architect |
| Performance | Fine for system config | Overkill for system config |
| Migration cost | Zero (keep going) | Full rewrite |

## Outcome

Decision: **Keep REST+NATS.** The architecture is sound. Ship features, not
rewrites. gRPC may be considered later for inter-service communication only.
