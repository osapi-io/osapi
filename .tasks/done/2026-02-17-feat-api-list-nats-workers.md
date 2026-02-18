---
title: API endpoint to list registered NATS workers
status: done
created: 2026-02-17
updated: 2026-02-17
---

## Objective

Add a REST API endpoint that lists which workers are currently
registered/connected to NATS. This gives operators visibility into
the fleet — which hosts are online, what they're subscribed to, and
their availability.

## Tasks

- [ ] Determine data source (NATS server monitoring API, consumer info,
      or heartbeat-based registration in KV)
- [ ] Design API response schema (hostname, subscriptions, connected
      since, last seen, etc.)
- [ ] Add OpenAPI spec for endpoint (e.g., `GET /workers` or
      `GET /job/workers`)
- [ ] Implement handler
- [ ] Add CLI command (`osapi client worker list` or similar)
- [ ] Add tests

## Design Considerations

- **NATS server monitoring**: The NATS server exposes connection and
  subscription info via its monitoring port. Could query this directly.
- **Consumer-based**: Query JetStream consumer info to see active
  consumers on the JOBS stream.
- **Heartbeat/KV-based**: Workers could write periodic heartbeats to a
  KV bucket (e.g., `workers.{hostname}` with TTL). This is the most
  reliable approach for knowing who is actually alive.
- The heartbeat approach would also support the `_all` broadcast use
  case — knowing how many workers to expect responses from.

## Notes

- Related to the CLI host targeting task — knowing which hosts exist
  helps users decide what to target
- Could be used by a future health endpoint to report worker
  availability
- The nats-server sibling repo may already expose some of this info
