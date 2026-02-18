---
title: "Update Docusaurus docs for job system migration"
status: done
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Ensure all Docusaurus documentation is updated to reflect the job system
migration. This is cross-cutting across all phases and should be done
after the code changes are complete (Phase 5).

## CLI Documentation

### Delete (task CLI docs no longer needed)
- `docs/docs/sidebar/usage/cli/task/` — task server, worker docs
- `docs/docs/sidebar/usage/cli/client/task/` — client task add, list,
  get, delete, status docs

### Add/Update (job CLI docs)
- `docs/docs/sidebar/usage/cli/client/job/add.md`
- `docs/docs/sidebar/usage/cli/client/job/list.md`
- `docs/docs/sidebar/usage/cli/client/job/get.md`
- `docs/docs/sidebar/usage/cli/client/job/delete.md`
- `docs/docs/sidebar/usage/cli/client/job/status.md`
- `docs/docs/sidebar/usage/cli/client/job/run.md`
- `docs/docs/sidebar/usage/cli/job/worker/start.md`

### Update sidebar config
- Update `sidebars.js` or equivalent to reference new job CLI paths

## API Documentation

### Update endpoint docs
- Update API cards/endpoints to reflect:
  - System API now routes through job client
  - Network API now uses strict-server with BearerAuth
  - Job API endpoints (POST /job, GET /job, etc.)
  - Task API endpoints removed

### Update architecture docs
- `docs/docs/sidebar/architecture.md`:
  - Mark migration phases 3-6 as complete
  - Update `OperationNetworkPingExecute` → `OperationNetworkPingDo`
  - Remove references to legacy task system
  - Update package architecture diagram

## Cards / Landing Pages

- Review any feature cards or landing pages that reference tasks
- Update to reference jobs instead
- Ensure consistent terminology (job, not task) throughout

## Verify

```bash
just docs::build  # Ensure docs build without errors
just docs::fmt-check  # Check formatting
grep -r "task" docs/docs/ --include="*.md" | grep -v node_modules
# Review results for stale task references
```

## Notes

This task depends on Phase 5 (delete legacy task) being complete so we
know exactly what needs to be removed from docs.

## Outcome

All Docusaurus documentation updated for job system migration:

- Deleted stale task CLI docs (6 client + 5 server/worker files)
- Deleted 6 generated task API .mdx docs
- Created 10 new CLI doc files: 7 client job commands + 3 job worker commands
- Regenerated OpenAPI merged spec via `redocly join` (includes job, excludes task)
- Regenerated Docusaurus API docs (`gen-api-docs all`) — sidebar now shows Job endpoints
- Fixed client code for regenerated types (`*string` error fields, `string` interface_name)
- Updated justfile: fetch downloads both `.mod.just` and `.just` files, uses `mod?` for optional loading
- Docusaurus build passes with no broken links
- All Go tests pass
