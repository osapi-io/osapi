---
title: Audit and clean up mocks repo-wide
status: done
created: 2026-02-17
updated: 2026-02-18
---

## Objective

Do a repo-wide audit of mocks with a focus on removing unused mocks and
eliminating hand-rolled assertion-based mocks when mockgen-generated mocks
already exist. The provider packages have many hand-written mocks in `mocks.go`
files that may be unnecessary or could be replaced with mockgen or inline
test-table mocks.

## Notes

### Mockgen-generated mocks (15 files, 8 directories) -- all actively used

| Directory                                              | Mocks                                                                                                  | Used By                            |
| ------------------------------------------------------ | ------------------------------------------------------------------------------------------------------ | ---------------------------------- |
| `internal/job/mocks/`                                  | NATSClient, KeyValue, KeyValueEntry, KeyWatcher, JobClient, NATSConnector, JetStream, JetStreamContext | job/client, job/worker, api/ tests |
| `internal/exec/mocks/`                                 | Manager                                                                                                | provider/network/dns tests         |
| `internal/provider/node/{host,disk,mem,load}/mocks/` | Provider (per domain)                                                                                  | job/worker tests                   |
| `internal/provider/network/{dns,ping}/mocks/`          | Provider, Pinger (ping only)                                                                           | job/worker, provider tests         |

### Hand-written wrapper files (`mocks.go`) -- 7 files, all useful

These provide `NewPlainMockProvider()` / `NewDefaultMockProvider()` factory
functions that configure mockgen-generated mocks with sensible defaults. Good
pattern -- kept all of them.

### Hand-rolled mocks -- 2 types, both appropriate to keep

1. **`mockHostnameProvider`** in `internal/job/hostname_test.go` -- simple stub
   for 1-method internal interface, no mockgen needed
2. **`mockJetStreamMsg`** in `internal/job/worker/consumer_test.go` -- minimal
   stub for 12-method external interface, only 2 methods matter, mockgen would
   add noise

### Unused mocks -- none found

### Assertion-based mocks -- none found

## Outcome

Audit complete. The codebase is clean:

- All 15 mockgen-generated mock files across 8 directories are actively
  referenced by tests
- All 7 hand-written `mocks.go` wrapper files provide useful factory functions
  and follow a consistent pattern
- The 2 hand-rolled mocks are justified (simple stubs where mockgen would be
  overkill)
- No unused mocks, no assertion-based mocks, no inconsistencies found
- No code changes required
