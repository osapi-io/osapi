---
title: Rename cmd/util_validate.go
status: backlog
created: 2026-02-17
updated: 2026-02-17
---

## Objective

`cmd/util_validate.go` needs to be renamed. The current name doesn't
follow the project's naming conventions for cmd/ files which use the
pattern `client_*.go`, `api_*.go`, `worker_*.go`, etc.

## Tasks

- [ ] Determine correct name based on what the file does and where it
      fits in the command hierarchy
- [ ] Rename the file
- [ ] Update any references/imports
- [ ] Verify tests still pass
