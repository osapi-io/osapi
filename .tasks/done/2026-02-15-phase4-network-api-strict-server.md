---
title: "Phase 4: Port network API to strict-server + job client"
status: done
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Rewrite network API from non-strict + direct providers + legacy tasks to
strict-server + job client. Biggest handler rewrite.

## Changes

- Enabled strict-server in `network/gen/cfg.yaml`
- Added BearerAuth security to all network OpenAPI endpoints
- Preserved `x-oapi-codegen-extra-tags` validation tags on all schemas
- Preserved 400 response types for validation errors
- Rewrote `types.go`: replaced PingProvider, DNSProvider, TaskClientManager
  with single `JobClient`
- Rewrote `network.go`: StrictServerInterface compile-time check
- Rewrote `network_dns_get_by_interface.go`: uses `JobClient.QueryNetworkDNS`
- Rewrote `network_dns_put_by_interface.go`: uses
  `JobClient.ModifyNetworkDNSAny`, keeps validator for request body
- Rewrote `network_ping_post.go`: uses `JobClient.QueryNetworkPingAny`,
  keeps validator for request body, keeps `durationToString` helper
- Rewrote `handler_network.go`: same pattern as system/job handlers with
  `NewStrictHandler` + `scopeMiddleware`
- Simplified `handler.go`: removed `appFs afero.Fs` parameter
- Simplified `manager.go`: interface matches new signature
- Updated `cmd/api_server_start.go`: removed `appFs` from CreateHandlers call

## Tests

- Deleted old integration tests (3 files)
- Created 3 new public test files with MockJobClient:
  - `network_dns_get_by_interface_public_test.go` (success, error)
  - `network_dns_put_by_interface_public_test.go` (success, validation error,
    error)
  - `network_ping_post_public_test.go` (success, validation error, error)
- Kept `network_test.go` (internal `durationToString` tests)
- All 11 tests pass

## Outcome

All network API handlers now use job client via NATS. All three API domains
(system, network, job) consistently use strict-server with BearerAuth JWT
middleware.
