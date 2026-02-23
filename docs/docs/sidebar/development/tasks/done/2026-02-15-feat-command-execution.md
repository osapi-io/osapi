---
title: Ad-hoc command execution
status: done
created: 2026-02-15
updated: 2026-02-22
---

## Objective

Add controlled command execution. Ansible's `command`, `shell`, and `raw`
modules are foundational — they allow running arbitrary commands when no
purpose-built module exists. An appliance API needs this as a fallback for
operations not covered by specific endpoints.

## API Endpoints

```
POST   /command/exec         - Execute command (no shell, no pipes)
POST   /command/shell        - Execute via shell (supports pipes, etc.)
```

## Operations

- `command.exec.execute` (modify)
- `command.shell.execute` (modify)

## Request Body

```json
{
  "command": "whoami",
  "args": [],
  "cwd": "/tmp",
  "timeout": 30,
  "user": "root"
}
```

## Response

```json
{
  "stdout": "root\n",
  "stderr": "",
  "exit_code": 0,
  "duration_ms": 12
}
```

## Provider

- `internal/provider/system/command/`
- Use existing `cmdexec` package for execution
- `exec` mode: no shell interpretation (safer, like Ansible `command`)
- `shell` mode: runs through `/bin/sh -c` (like Ansible `shell`)
- Support timeout, working directory, run-as user

## Notes

- This is the most security-sensitive feature — needs careful controls
- Consider command allowlisting or denylisting
- Require highest privilege scope: `command:execute`
- Log all executions to audit log
- Set hard timeout limits (max 5 minutes?)
- The existing `cmdexec.Manager` already provides safe execution
- Long-running commands are a natural fit for the async job system

## Outcome

Implemented the full command execution domain:

- **Provider**: `internal/provider/command/` with `Exec()` and `Shell()` methods
  using the new `exec.RunCmdFull()` for separate stdout/stderr
- **OpenAPI**: Two POST endpoints (`/command/exec`, `/command/shell`) with
  validation, auth, and 202 async responses
- **Job system**: Operation constants, data structs, job client methods
  (single + broadcast), and worker processor dispatch
- **Permissions**: `command:execute` permission, admin-only by default
- **CLI**: `osapi client command exec` and `osapi client command shell` with
  `--command`, `--args`, `--cwd`, `--timeout`, `--target`, `--json`
- **Tests**: 22 new tests (unit + integration + RBAC), all passing
- **Docs**: Feature page, CLI pages, config/permission updates
- `run-as user` deferred (requires syscall.SysProcAttr, linux-only)
