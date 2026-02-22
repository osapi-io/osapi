---
title: Audit CLI commands and docs for consistency
status: done
created: 2026-02-18
updated: 2026-02-18
---

## Objective

Audit all CLI commands and ensure documentation is up to date and consistent
across every command. After the label-based routing feature, several docs were
updated but there may be inconsistencies remaining.

## Scope

- Review every `cmd/client_*.go` file for flag descriptions, defaults, and help
  text
- Cross-reference each CLI command against its corresponding doc in
  `docs/docs/sidebar/usage/cli/client/`
- Verify flag tables, example output, and targeting examples are consistent
- Ensure `--target` flag description and examples use the new hierarchical label
  syntax (`group:web.dev`) everywhere
- Check that OpenAPI spec parameter descriptions match CLI help text
- Verify the architecture docs accurately describe all supported target types

## Notes

- The label-based routing feature changed target syntax and added hierarchical
  labels — make sure all references use the new format
- Check for any stale references to old flat label syntax (e.g., `role:web`)

## Outcome

Added `-T, --target` flag documentation to 5 CLI doc files:

- `system/hostname.md` — added new Flags section
- `system/status.md` — added new Flags section
- `network/dns/get.md` — added row to existing Flags table
- `network/dns/update.md` — added row to existing Flags table
- `network/ping/ping.md` — added row to existing Flags table

No stale label syntax found; all examples already use hierarchical format.
