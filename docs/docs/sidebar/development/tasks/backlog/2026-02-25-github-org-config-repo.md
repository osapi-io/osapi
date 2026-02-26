---
title: Extract GitHub org config into dedicated repo
status: backlog
created: 2026-02-25
updated: 2026-02-25
---

## Objective

Move the `github/` directory (declarative repo settings and sync tooling) out of
`osapi` into a dedicated `osapi-io/.github` or `osapi-io/github-config` repo.

## Context

The `github/` directory currently lives untracked in the main `osapi` repo. It
contains:

- `repos.json` — declarative GitHub org config (repo settings, branch protection
  rules)
- `sync.sh` — Bash script that uses `gh` CLI to detect drift and apply config

This tooling manages org-wide settings across all `osapi-io` repos, so it
doesn't belong inside a single project repo.

## Research

Look into existing tools for declarative GitHub org/repo management:

- [Terraform GitHub provider](https://registry.terraform.io/providers/integrations/github/latest)
- [github-mgmt](https://github.com/toptal/github-mgmt) — GitOps for GitHub org
  settings
- [Probot settings](https://github.com/probot/settings) — `.github/settings.yml`
  applied via GitHub App
- [safe-settings](https://github.com/github/safe-settings) — GitHub's own
  org-level settings management

Evaluate whether an off-the-shelf solution fits before building custom tooling.

## Notes

- The current `sync.sh` script is functional but minimal
- Branch protection and repo settings are already configured via GitHub UI / API
  — the script checks for drift
- A dedicated repo would allow CI to run the sync on a schedule or on push
