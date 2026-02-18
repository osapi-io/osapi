---
title: "Feature: File and directory management"
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add file and directory management. Ansible's `file`, `copy`, `template`,
`lineinfile`, and `stat` modules are among its most popular — operators
constantly need to manage config files, set permissions, and inspect
file state on remote systems.

## API Endpoints

```
GET    /file/stat             - Get file/directory metadata
POST   /file/read             - Read file contents (with line range)
PUT    /file/write            - Write/create file with content
PATCH  /file/line             - Insert/replace/remove line in file
PUT    /file/permissions      - Set owner, group, mode
POST   /file/directory        - Create directory (mkdir -p)
DELETE /file/{path}           - Delete file or directory

POST   /file/copy             - Copy file within the system
POST   /file/archive          - Create tar/gz archive
POST   /file/extract          - Extract archive
```

## Operations

- `file.stat.get` (query) — Ansible `stat` equivalent
- `file.read.get` (query)
- `file.write.execute` (modify) — Ansible `copy`/`template` equivalent
- `file.line.update` (modify) — Ansible `lineinfile` equivalent
- `file.permissions.update` (modify) — Ansible `file` with mode/owner
- `file.directory.create` (modify)
- `file.delete.execute` (modify)
- `file.copy.execute` (modify)
- `file.archive.create`, `file.archive.extract` (modify)

## Provider

- `internal/provider/system/file/`
- `stat`: Return type with path, size, mode, owner, group, modified,
  is_dir, is_link, checksum (sha256)
- `read`: Support line offset/limit, binary detection
- `write`: Accept content + mode + owner, create parent dirs
- `line`: Support regexp match, insertafter, insertbefore, state
  (present/absent) — mirrors Ansible lineinfile semantics
- `permissions`: `chown`, `chmod`

## Notes

- Ansible's `lineinfile` is one of its most-used modules — the
  regex-based line management is very powerful for config editing
- File write should support backup (rename original before overwrite)
- Size limits on read/write to prevent abuse
- Path validation to prevent directory traversal attacks
- Scopes: `file:read`, `file:write`
- Consider a diff endpoint to preview changes before applying
