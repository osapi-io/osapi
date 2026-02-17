---
title: Add Darwin development providers
status: done
created: 2026-02-16
updated: 2026-02-16
---

## Objective

Add Darwin providers so the job worker runs on macOS without errors. Use real
gopsutil data for system providers (host, disk, mem, load) and mock data for
DNS and ping since they depend on platform-specific tooling.

## Notes

- Branch: `feat/darwin-providers`
- System providers use same gopsutil calls as Ubuntu (cross-platform)
- Network providers return mock data (DNS, ping)
- Remove darwin hacks from Linux DNS provider
- Log warning at startup when running on darwin

## Outcome

Implemented 24 new Darwin provider files, modified 7 existing files. All
tests pass, 0 lint issues. Committed on `feat/darwin-providers` branch.
