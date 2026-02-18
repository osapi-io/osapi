---
title: Fix dotted hostname breaks NATS subject routing
status: done
created: 2026-02-18
updated: 2026-02-18
---

## Objective

Hostnames containing dots (e.g., `Johns-MacBook-Pro-2.local`) break NATS
subject routing because dots are NATS subject delimiters.

## Problem

When targeting a host with dots in the name, `BuildSubjectFromTarget` produces
a subject like `jobs.query.host.Johns-MacBook-Pro-2.local` (6 tokens). The
worker subscribes using `SanitizeHostname`, which replaces dots with
underscores: `jobs.*.host.Johns-MacBook-Pro-2_local` (4 tokens). The subjects
never match, so the job is never delivered.

## Affected Code

- `internal/job/subjects.go` — `BuildSubjectFromTarget` (line 241) builds the
  publish subject without sanitizing the hostname
- `internal/job/subjects.go` — `BuildWorkerSubscriptionPattern` (line 148)
  sanitizes via `SanitizeHostname`
- `internal/job/subjects.go` — `ParseSubject` (line 113) would also need to
  handle sanitized hostnames on the parse side

## Suggested Fix

Sanitize the hostname in `BuildSubjectFromTarget` for the `"host"` case so
the publish subject matches the worker subscription. Ensure `ParseSubject`
round-trips correctly with sanitized hostnames. Add test cases for dotted
hostnames.

## Notes

- Reproduced with: `go run main.go client network dns get --interface-name eth0 --target 'Johns-MacBook-Pro-2.local'`
- The job is created but no worker ever picks it up.
