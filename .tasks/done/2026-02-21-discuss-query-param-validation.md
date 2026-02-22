---
title: Evaluate query param validation strategy
status: done
created: 2026-02-21
updated: 2026-02-22
---

## Objective

Decide whether to keep manual query parameter validation or adopt a
different approach.

## Context

oapi-codegen generates `validate:` struct tags from
`x-oapi-codegen-extra-tags` on **request body properties** but NOT on
**query parameters** when the extension is nested inside `schema:`.

## Outcome

Discovered that placing `x-oapi-codegen-extra-tags` at the **parameter
level** (sibling of `name`/`in`/`schema`) instead of inside `schema:`
causes oapi-codegen to generate `validate:` tags on query param struct
fields. This eliminates the need for manual temporary-struct validation.

Changes made:
- Moved `x-oapi-codegen-extra-tags` to parameter level in all four
  OpenAPI specs (audit, job, system, network)
- Replaced manual temp-struct validation with
  `validation.Struct(request.Params)` in all seven affected handlers
- Added "when empty target_hostname returns 400" integration test cases
  to all five endpoints that accept `target_hostname`
- Updated CLAUDE.md to document the correct placement
- All tests pass, 0 lint issues

Note: `x-oapi-codegen-extra-tags` on **path parameters** does NOT
generate tags on request object structs. Path params still need manual
validation for non-UUID types (e.g., `interfaceName` alphanum check).
