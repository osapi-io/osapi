---
title: Replace box-drawing tables with compact kubectl-style columns
status: done
created: 2026-02-26
updated: 2026-02-26
---

## Objective

Replace heavy box-drawing tables (`PrintStyledTable`) with compact
column-aligned output (`PrintCompactTable`) across all CLI list and broadcast
views. Aligns CLI output with kubectl conventions: no borders, colored headers,
aligned columns, truncation for long values.

## Changes

- Added `PrintCompactTable` to `internal/cli/ui.go` — renders kubectl-style
  compact tables with purple headers, teal data, 2-space indent, 2-space column
  gaps
- Multi-line cell values flattened to single lines (`strings.Fields` join)
- Long values truncated at 50 characters with ellipsis
- Removed `PrintStyledTable`, `lipgloss/table` import, `os` import
- Migrated all 12 call sites across 10 cmd/ files
- Updated `TestPrintStyledTable` → `TestPrintCompactTable` with header content
  assertions
- Updated all 14 CLI doc files to show compact output instead of box-drawing
  tables
- Verified flags tables in docs match actual `--help` output

## Outcome

Three consistent output patterns across every CLI command:

1. **Detail views** — `PrintKV` (node get, audit get, health status)
2. **List views** — `PrintCompactTable` (node list, job list, audit list)
3. **Broadcast results** — `PrintCompactTable` (hostname, status, dns, ping,
   exec, shell)

Full data always available via `--json` for values truncated in compact view.
