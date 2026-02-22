---
title: Rename cmd/util_*.go files to descriptive names
status: done
created: 2026-02-17
updated: 2026-02-17
---

## Objective

`cmd/util_validate.go` needs to be renamed. The current name doesn't follow the
project's naming conventions for cmd/ files which use the pattern `client_*.go`,
`api_*.go`, `worker_*.go`, etc.

## Tasks

- [x] Determine correct name based on what the file does and where it fits in
      the command hierarchy
- [x] Rename the file
- [x] Update any references/imports
- [x] Verify tests still pass

## Outcome

Renamed all three `util_` prefixed files to descriptive names, matching the
convention used by major Cobra CLI projects (Docker CLI, NATS CLI):

- `cmd/util_validate.go` → `cmd/validate.go`
- `cmd/util_log.go` → `cmd/log.go`
- `cmd/util_ui.go` → `cmd/ui.go`

No import changes needed — all files are in `package cmd`. Build, tests, and vet
all pass.
