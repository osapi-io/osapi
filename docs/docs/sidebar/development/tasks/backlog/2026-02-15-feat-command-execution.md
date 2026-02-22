---
title: Ad-hoc command execution
status: backlog
created: 2026-02-15
updated: 2026-02-15
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
