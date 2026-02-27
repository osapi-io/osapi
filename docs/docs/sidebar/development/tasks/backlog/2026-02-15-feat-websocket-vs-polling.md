---
title: Evaluate WebSocket vs polling for job status updates
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Evaluate whether to add WebSocket support for real-time job status updates as an
alternative to the current HTTP polling approach. Determine the right strategy
and add it to the implementation roadmap.

## Current State

The REST API currently supports polling for job status:

- `GET /job/{id}` — poll for individual job status
- `GET /job/status` — poll for queue statistics
- The CLI `client job status` uses BubbleTea with configurable poll interval

## Options to Evaluate

### Option 1: WebSocket for Real-Time Status

- `WS /job/{id}/ws` — stream status events for a specific job
- `WS /job/status/ws` — stream queue statistics in real time
- Pros: Low latency, no wasted polls, better UX
- Cons: More complex server code, connection management, auth per-connection

### Option 2: Server-Sent Events (SSE)

- `GET /job/{id}/events` — SSE stream for job status
- Pros: Simpler than WebSocket, works with HTTP/2, one-directional
- Cons: No bidirectional communication, browser support varies

### Option 3: Keep Polling, Optimize

- Add `ETag`/`If-None-Match` for conditional responses
- Add long-polling with `?wait=30s` parameter
- Pros: Simplest, REST-compatible
- Cons: Higher latency, more server load

## Considerations

- NATS JetStream already provides real-time notifications internally — the
  question is whether to expose this to REST clients
- WebSocket would align well with the `_all` broadcast jobs where status updates
  come from multiple agents over time
- Echo framework has WebSocket support via `labstack/echo`
- Auth needs to work for long-lived connections (token refresh, expiry)
- Consider whether the CLI and API need the same approach or can differ

## Decision Criteria

- Expected client usage patterns (dashboards? CI/CD? one-shot scripts?)
- Complexity budget vs user experience gain
- Whether polling with reasonable intervals (2-5s) is "good enough"
