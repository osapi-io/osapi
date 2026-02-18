---
title: "Phase 6: Consistency pass and test coverage audit"
status: done
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Final cleanup to ensure everything is uniform. Verify auth consistency,
error response consistency, and test coverage.

## Changes

- Verify all domains use scopeMiddleware with BearerAuth
- Evaluate error response unification
- Audit validation on all API endpoints (see
  `2026-02-15-api-endpoint-validation-audit.md`)
- Test coverage audit (target near-100% on non-generated, non-CLI code)
- Add middleware tests if not present
- Update architecture.md migration path (mark phases complete)
- Remove task system references from architecture.md
- Regenerate combined OpenAPI spec
- Review all docs for stale task-system references

## Verify

```bash
just test
go test -cover ./internal/api/...
go test -cover ./internal/job/...
```

## Outcome

Completed. Key changes:

### Auth Consistency
- All 3 domains (system, network, job) use `scopeMiddleware` consistently
- Added missing `securitySchemes` to system API OpenAPI spec
- **Fixed bug**: Network API used `network:read`/`network:write` scopes
  which don't exist in `RoleHierarchy` — would always return 403. Changed
  to `read`/`write` to match system and job domains.
- Added 401/403 response types to all network API endpoints

### Error Response Consistency
- Replaced network's local `network.ErrorResponse` with common
  `ErrorResponse` reference (matching system and job)
- Added `import-mapping` for common API to network `cfg.yaml`
- Updated network handlers to use pointer fields (`&errMsg`) matching
  the common ErrorResponse type

### Test Coverage
- Created `internal/api/middleware_test.go` with 19 tests covering
  `scopeMiddleware` and `hasScope` (11 hasScope + 8 scopeMiddleware cases)
- Improved job API from 75.8% to **100%** by adding optional field coverage
  tests for GetJobByID and GetJob (list)
- Final coverage: api/job 100%, api/network 100%, api/system 100%

### Documentation
- Updated CLAUDE.md: removed task references from architecture section
- Updated principles.md: "Task Worker" → "Job System"
- Updated architecture.md: marked all 6 migration phases complete
- Updated common/gen/api.yaml: "Task Worker" → "Job System"
- Deleted stale task CLI docs (6 client task + 5 server/worker task files)
- Deleted 6 generated task API .mdx docs

### Copyright
- Updated all new files from 2025 → 2026 copyright year
