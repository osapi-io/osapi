---
title: Evaluate synadia-io/orbit.go for NATS integration
status: done
created: 2026-02-17
updated: 2026-02-18
---

## Objective

Evaluate whether https://github.com/synadia-io/orbit.go would be a good
fit for OSAPI's NATS JetStream integration, potentially replacing or
complementing the current nats-client sibling repo.

## Evaluation

### orbit.go Overview

orbit.go is a collection of **independent Go modules** (separate `go.mod`
each) from Synadia that provide higher-level abstractions over the NATS Go
client. All modules are **pre-v1.0** with no API stability guarantees. The
repo has 68 stars, 8 contributors, and active development (11 releases as
of Feb 2026).

### Module-by-Module Assessment

#### jetstreamext — JetStream Extensions (HIGH relevance)

Provides `GetBatch` (fetch N messages from a stream by subject/sequence/
time) and `GetLastMsgsFor` (get last message for each of N subjects).
Also adds atomic `PublishMsgBatch` for all-or-nothing multi-message
publishing.

- `GetBatch` could improve job listing by fetching stream messages with
  subject filters instead of scanning all KV keys.
- `GetLastMsgsFor` maps directly to "get latest status event per job"
  queries.
- Batch publish enables atomic multi-job creation if needed later.
- **Caveat:** Batch publish requires nats-server v2.12.0+. Requires Go
  1.24+ (OSAPI uses 1.25, so fine).

#### kvcodec — KV Codecs (MODERATE-HIGH relevance)

Wraps `jetstream.KeyValue` to transparently encode/decode keys and values.
Includes `Base64Codec` and `PathCodec` (converts `/` paths to `.` NATS
subjects). Implements the full `jetstream.KeyValue` interface — drop-in
wrapper.

- OSAPI already solved the dotted-hostname problem by sanitizing dots in
  `BuildSubjectFromTarget`. kvcodec would be a cleaner, more general
  solution.
- Zero changes to existing KV operations — just wrap the bucket.
- Useful insurance against future key-encoding edge cases with labels or
  hostnames.

#### natsext — Core NATS Extensions (MODERATE-HIGH relevance)

Adds `RequestMany` / `RequestManyMsg` for scatter-gather patterns: send
one request, collect multiple replies with stall timers, max message caps,
and sentinel detection.

- OSAPI's broadcast response collection currently uses KV watch + quiet
  period polling (3-second silence window). `RequestMany` would enable
  real-time scatter-gather without the KV intermediary.
- Most useful if OSAPI adds a synchronous "query all workers and aggregate
  in real time" mode.

#### pcgroups — Partitioned Consumer Groups (MODERATE relevance)

Partitions messages by key so same-key messages always go to the same
consumer, while different keys process in parallel. Both static and
elastic (dynamic membership) modes.

- OSAPI's current queue group model works well for `_any` routing. PCGroups
  would matter if we needed per-resource ordering guarantees (e.g.,
  "all DNS updates for eth0 go to the same worker to avoid conflicts").
- Requires nats-server 2.11+.

#### natssysclient — NATS System Client (LOW-MODERATE relevance)

Type-safe access to NATS monitoring APIs (`varz`, `connz`, `jsz`,
`healthz`, etc.) with single-server and cluster-wide scatter-gather.

- Useful for observability dashboards but not for core job processing.
- Could complement a future monitoring/health feature.

#### counters — Distributed Counters (LOW-MODERATE relevance)

Arbitrary-precision distributed counters backed by JetStream streams.

- Could track "total jobs processed per worker" or "failures per host"
  without a separate metrics database.
- OSAPI computes queue stats by scanning KV keys today, which works fine.

#### natscontext — NATS Context (LOW relevance)

Connects using NATS CLI saved context profiles. OSAPI manages its own
connections through `osapi.yaml` / Viper. Not applicable.

### Comparison with Current nats-client

OSAPI wraps `github.com/osapi-io/nats-client` behind a
`messaging.NATSClient` interface with methods for: `Connect`,
`CreateOrUpdateStreamWithConfig`, `CreateOrUpdateConsumerWithConfig`,
`CreateKVBucket`, `KVPut/Get/Delete/Keys`, `Publish`,
`PublishAndWaitKV`, `ConsumeMessages`, `GetStreamInfo`.

| Concern | nats-client | orbit.go |
|---------|-------------|----------|
| **Scope** | Full NATS abstraction (connect, streams, KV, consumers, publish, consume) | Individual utilities that enhance specific patterns |
| **Architecture** | Single wrapper with unified interface | Independent modules, each solving one problem |
| **Replaceability** | orbit.go cannot replace nats-client — it has no connection management, stream/consumer creation, or basic KV CRUD | orbit.go modules layer on top of raw nats.go, same level as nats-client |
| **Testing** | OSAPI mocks `messaging.NATSClient` for all tests | Each orbit module is independently testable |

**Key insight:** orbit.go is not a replacement for nats-client. It's a
collection of enhancements that could complement either nats-client or
raw nats.go usage.

### Maturity & Risk

- All modules are pre-v1.0 — APIs may break between releases.
- Small community (68 stars, 8 contributors).
- Backed by Synadia (the company behind NATS), so unlikely to be
  abandoned, but may evolve rapidly.
- Some modules require very recent nats-server versions (2.11+, 2.12+).

## Recommendation: Skip (revisit when modules hit v1.0)

**Do not adopt orbit.go modules now.** The risk/reward doesn't justify it:

1. **nats-client works well.** OSAPI's current abstraction covers all
   needs. The `messaging.NATSClient` interface is clean and testable.

2. **Pre-v1.0 instability.** Adopting pre-v1.0 modules means accepting
   breaking changes on every upgrade. OSAPI's NATS integration is
   foundational — instability here is costly.

3. **No critical gap.** The most relevant modules (jetstreamext, kvcodec,
   natsext) solve optimization problems, not missing functionality.
   OSAPI's KV-scan approach works, the dot-sanitization fix works, and
   KV watch polling works.

4. **Dependency weight.** Each orbit module pulls its own dependency tree.
   Adding multiple modules increases maintenance surface for marginal
   gains.

**When to revisit:**

- When `jetstreamext` or `kvcodec` reach v1.0 — these are the most
  useful modules and stable APIs would make adoption worthwhile.
- If KV-scan performance becomes a bottleneck at scale — `GetBatch` and
  `GetLastMsgsFor` would be the right solution.
- If OSAPI adds real-time scatter-gather queries — `RequestMany` from
  natsext would be cleaner than the current KV watch approach.

## Outcome

Completed evaluation. Recommendation is **skip for now** — orbit.go
modules are interesting but pre-v1.0 and solve optimization problems
rather than missing functionality. The most relevant modules
(jetstreamext, kvcodec, natsext) should be revisited when they stabilize.
