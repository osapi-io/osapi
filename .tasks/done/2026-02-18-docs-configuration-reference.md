---
title: Add configuration reference documentation
status: done
created: 2026-02-18
updated: 2026-02-18
---

## Objective

Document the `osapi.yaml` config file structure and `OSAPI_` environment
variable overrides. Users currently have no reference for available configuration
options outside of reading the Go source.

## Scope

- Document all config sections from `internal/config/types.go`: API (client,
  server, security), NATS (server), Job (stream, consumer, KV, DLQ, client,
  worker)
- Show example `osapi.yaml` with all fields and defaults
- Document the `OSAPI_` env var convention (`AutomaticEnv` + `SetEnvPrefix`)
  with key mapping examples (dots become underscores, e.g.,
  `api.server.port` → `OSAPI_API_SERVER_PORT`)
- Note which fields are required vs optional and which are sensitive
  (`signing_key`, `bearer_token`)

## Notes

- Viper config is in `cmd/root.go` (`initConfig`)
- Config types are in `internal/config/types.go`
- Validation is in `internal/config/schema.go`
- The `--osapi-file` / `-f` flag controls config file path (default `osapi.yaml`)

## Outcome

- Created `docs/docs/sidebar/configuration.md` with full annotated YAML
  reference, env var mapping table, required fields, and per-section tables
- Fixed `system-architecture.md` config skeleton to nest `url` and
  `bearer_token` under `api.client` instead of directly under `api`
- Added cross-link from system-architecture to the new configuration page
