---
title: Add NATS authentication and namespace support to config
status: backlog
created: 2026-02-20
updated: 2026-02-20
---

## Objective

The nats-client library already supports three auth types (NoAuth, UserPass,
NKey), but OSAPI hardcodes `NoAuth` in all three components. Surface these
auth options through `osapi.yaml` and plumb them into the client connections.
Also add namespace/subject prefix support so multiple OSAPI deployments can
share a NATS cluster without colliding.

## Current State

- `nats-client/pkg/client/types.go` defines `AuthType` enum: `NoAuth`,
  `UserPassAuth`, `NKeyAuth`
- `nats-client/pkg/client/connect.go` implements all three auth flows
- All OSAPI startup commands hardcode `natsclient.NoAuth`:
  - `cmd/api_server_start.go`
  - `cmd/job_worker_start.go`
  - `cmd/nats_server_start.go`
- No auth fields exist in `internal/config/types.go` for NATS
- The embedded NATS server has no auth configured server-side

## Proposed Config

```yaml
nats:
  server:
    host: 0.0.0.0
    port: 4222
    store_dir: .nats/jetstream/
    auth:
      # Auth type for the embedded server: "none", "user_pass", "nkey"
      type: none
      # For user_pass: configure allowed users
      users:
        - username: osapi
          password: '<secret>'
      # For nkey: configure allowed NKeys
      # nkeys:
      #   - '<public-nkey>'

api:
  server:
    nats:
      host: localhost
      port: 4222
      client_name: osapi-api
      auth:
        type: none          # "none", "user_pass", or "nkey"
        username: ''        # for user_pass
        password: ''        # for user_pass
        nkey_file: ''       # for nkey (path to seed file)
      namespace: ''         # subject prefix, e.g., "prod" -> "prod.jobs.>"

job:
  worker:
    nats:
      host: localhost
      port: 4222
      client_name: osapi-job-worker
      auth:
        type: none
        username: ''
        password: ''
        nkey_file: ''
      namespace: ''
```

## Changes Required

### Config Layer

- `internal/config/types.go` — add `NATSAuth` struct with `Type`,
  `Username`, `Password`, `NKeyFile` fields; add `Namespace` field;
  embed in NATS connection sections
- `osapi.yaml` — add auth and namespace examples
- `docs/docs/sidebar/configuration.md` — document new fields
- Environment variable mappings (e.g., `OSAPI_NATS_SERVER_AUTH_TYPE`)

### Client Plumbing

- `cmd/api_server_start.go` — read auth config, build
  `natsclient.AuthOptions` from config instead of hardcoding `NoAuth`
- `cmd/job_worker_start.go` — same
- `cmd/nats_server_start.go` — configure server-side auth (check
  nats-server sibling repo for options)

### Namespace Support

- Prefix all NATS subjects (`jobs.>`) with configurable namespace
- Prefix stream names, KV bucket names, consumer names
- Update `internal/job/` subject routing to respect namespace
- Ensure multiple OSAPI deployments on same NATS cluster don't collide

### Documentation

- Document all three auth types with examples
- Document namespace usage for multi-tenant deployments
- Security recommendations (NKey preferred for production)

## Related

- `.tasks/backlog/2026-02-20-refactor-nats-config.md` — NATS config
  restructuring (should be done first or in conjunction)

## Notes

- This is a backwards-compatible change (defaults to `NoAuth` and no
  namespace)
- The nats-client library is already prepared — this is mostly config
  plumbing
- NKey auth examples exist in `nats-client/examples/auth-nkeys-stream/`
- UserPass auth examples in `nats-client/examples/auth-user-pass-stream/`

## Outcome

_To be filled in when done._
