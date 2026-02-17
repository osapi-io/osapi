---
title: Audit and standardize CLI documentation formatting
status: done
created: 2026-02-15
updated: 2026-02-17
---

## Objective

All CLI documentation pages under `docs/docs/sidebar/usage/cli/` should use
consistent formatting conventions. Audit all existing CLI docs and fix any
inconsistencies found.

### Formatting Rules

- All command examples use ` ```bash ` fenced code blocks
- All command prompts use `$` PS1 prefix
- Output examples are clearly separated from commands (no `$` prefix on output
  lines)
- Flag/option tables use consistent column headers and formatting
- All pages follow the same structural pattern (title, description, usage, flags
  table, examples)

## Scope

Audit and fix CLI documentation for:

- `api` command
- `client` command (including `job`, `system`, `network` subcommands)
- `job worker` command
- `nats-server` command

## Desired State

Every CLI doc page should follow this structural pattern:

1. **Title** - Command name as heading
2. **Description** - Brief explanation of what the command does
3. **Usage** - `bash` fenced code block with `$` prompt prefix showing
   invocation
4. **Flags table** - Consistent column headers (e.g., Flag, Type, Default,
   Description) with uniform formatting
5. **Examples** - `bash` fenced code blocks with `$` prompt on command lines
   and no `$` prefix on output lines

## Notes

- Ensure 80-character line wrap in all Markdown files (enforced by Prettier).
- Copyright year: 2026.
- Check for any pages that deviate from the structural pattern and align them.
