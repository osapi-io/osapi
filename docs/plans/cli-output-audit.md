# CLI Output Audit — 100-char Width Target

## Goal

Standardize all CLI table output to fit within ~100 characters. Merge
STATUS/CHANGED/ERROR into a single STATUS column. Show error details
in a separate section below the table.

## STATUS Column Values

| Value     | Meaning                                   |
| --------- | ----------------------------------------- |
| `ok`      | Succeeded, no change needed               |
| `changed` | Succeeded and modified system state       |
| `skip`    | Agent can't perform (unsupported OS)      |
| `err`     | Operation failed (details shown below)    |

## Table Format

```
  Job ID: ...

  HOSTNAME  STATUS   NAME      UID    HOME
  nerd      ok       root      0      /root
  nerd      ok       retr0h    1000   /home/retr0h
  mac       skip

  Errors:
  mac  operation not supported on this OS family
```

## Core Changes

### BuildBroadcastTable + BuildMutationTable

- Always show HOSTNAME + STATUS
- STATUS = `ok` | `changed` | `skip` | `err`
- Remove ERROR column from table
- Remove CHANGED column (merged into STATUS)
- Return error list separately for rendering below table

### PrintCompactTable

- Accept optional errors section
- Render errors below table when present

## Per-Command Column Audit

### Keep as-is (already compact)

| Command | Columns | Est. |
|---------|---------|------|
| hostname get | (none — just HOSTNAME + STATUS) | ~40 |
| uptime get | UPTIME | ~45 |
| os get | DISTRIBUTION, VERSION | ~55 |
| memory get | TOTAL, USED, FREE, USAGE | ~65 |
| load get | LOAD (1m), LOAD (5m), LOAD (15m) | ~65 |
| timezone get | TIMEZONE, UTC_OFFSET | ~55 |
| sysctl get/list | KEY, VALUE | ~60 |
| group get/list | NAME, GID, MEMBERS | ~60 |
| certificate list | NAME, SOURCE | ~55 |
| log source | SOURCE | ~45 |

### Trim columns

| Command | Current | Proposed | Dropped |
|---------|---------|----------|---------|
| hostname get | LABELS | (none) | LABELS (use node get) |
| user list | NAME, UID, GID, HOME, SHELL, GROUPS, LOCKED | NAME, UID, HOME, SHELL, GROUPS | GID, LOCKED |
| user get | NAME, UID, GID, HOME, SHELL, GROUPS, LOCKED | NAME, UID, HOME, SHELL, GROUPS | GID, LOCKED |
| process list | PID, NAME, USER, STATE, CPU%, MEM%, COMMAND | PID, NAME, USER, STATE, CPU%, COMMAND | MEM% |
| process get | PID, NAME, USER, STATE, CPU%, MEM%, COMMAND | PID, NAME, USER, STATE, CPU%, COMMAND | MEM% |
| service get | NAME, STATUS, ENABLED, DESCRIPTION, PID | NAME, STATUS, ENABLED, DESCRIPTION | PID |
| service list | NAME, STATUS, ENABLED, DESCRIPTION | NAME, STATUS, ENABLED | DESCRIPTION |
| cron list | NAME, SOURCE, SCHEDULE, OBJECT, USER | NAME, SCHEDULE, OBJECT, USER | SOURCE |
| ntp get | SYNCHRONIZED, STRATUM, OFFSET, SOURCE, SERVERS | SYNCHRONIZED, SOURCE, SERVERS | STRATUM, OFFSET |
| ping | AVG RTT, MIN RTT, MAX RTT, PACKET LOSS, PACKETS RECEIVED | AVG RTT, MIN RTT, MAX RTT, LOSS | PACKETS RECEIVED |
| docker list | ID, NAME, IMAGE, STATE, CREATED | NAME, IMAGE, STATE | ID, CREATED |
| docker inspect | ID, NAME, IMAGE, ... | NAME, IMAGE, STATE | (trim to essentials) |
| disk get | MOUNT, TOTAL, USED, FREE, USAGE | MOUNT, TOTAL, USED, USAGE | FREE |
| log query/unit | TIMESTAMP, PRIORITY, UNIT, MESSAGE | TIMESTAMP, UNIT, MESSAGE | PRIORITY |
| command exec/shell | STDOUT, STDERR, EXIT CODE, DURATION | EXIT CODE, STDOUT | STDERR (separate), DURATION |
| status get | UPTIME, LOAD (1m), MEMORY USED | UPTIME, LOAD, MEM | shorten headers |
| package list | NAME, VERSION, STATUS, SIZE | NAME, VERSION, STATUS | SIZE |
| file status | PATH, STATUS, SHA256 | PATH, STATUS | SHA256 (use --json) |
| ssh key list | TYPE, FINGERPRINT, COMMENT | TYPE, FINGERPRINT | COMMENT |

## Implementation Order

1. Rewrite `BuildBroadcastTable` — unified STATUS, separate errors
2. Rewrite `BuildMutationTable` — same pattern (or merge into one)
3. Update `PrintCompactTable` — render errors section
4. Update tests for new table format
5. Trim columns per command (one commit per domain group)
6. Update CLI docs
