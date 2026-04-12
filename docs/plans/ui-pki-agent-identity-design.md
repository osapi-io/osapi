# UI PKI & Agent Identity — Design Spec

Date: 2026-04-12

## Overview

Update the React management dashboard to show agent identity (machine
ID, fingerprint), PKI enrollment state (Pending), and missing job
status colors. Add a pending agents admin section with accept/reject
actions. Register PKI commands in the `:` command bar.

## 1. Agent Card Updates

**File:** `ui/src/components/domain/agent-card.tsx`

### Machine ID

Show truncated machine ID below the hostname. Use the existing `Text`
component with `variant="muted"` and a `title` attribute for the full
value on hover. Clicking copies the full machine ID to clipboard.

```
┌─────────────────────────────────────┐
│ 🖥 web-01                    Ready  │
│ a1b2c3d4-e5f6...           Pending  │
│ SHA256:4fee...                      │
│                                     │
│ CPU: 0.42  MEM: 4.2/8 GB  UP: 3d   │
│ Ubuntu 24.04 / amd64 / 4 cpu       │
│ ● facts ● heartbeat ● pki          │
│ ⚠ DiskPressure                      │
│ group:web.dev.us-east               │
│ [Drain]                             │
└─────────────────────────────────────┘
```

### Fingerprint

Show fingerprint when present (PKI enabled). Truncate to ~20 chars
with `...`. Use `Text variant="muted"`. Only shown when
`agent.fingerprint` is non-empty.

### Pending State

Add `Pending` to `stateVariant()`:
```typescript
case "Pending":
  return "pending" as const;
```

When state is `Pending`:
- Show `Pending` badge (yellow, uses existing `pending` variant)
- Hide drain/undrain buttons
- Show `Text variant="muted"`: "Awaiting PKI enrollment"

## 2. Job Status Colors

**File:** `ui/src/pages/jobs.tsx`

Update `statusBadgeVariant()` to handle all job statuses:

```typescript
function statusBadgeVariant(status?: string) {
  switch (status) {
    case "completed":
      return "ready" as const;
    case "failed":
    case "partial_failure":
      return "error" as const;
    case "processing":
    case "acknowledged":
    case "started":
      return "running" as const;
    case "submitted":
      return "pending" as const;
    case "skipped":
      return "muted" as const;
    case "retried":
      return "pending" as const;
    default:
      return "muted" as const;
  }
}
```

## 3. Pending Agents Section

### Location

Add a "Pending Agents" section to the Dashboard page, above or below
the existing agent cards. Only visible when there are pending agents
AND the user has `agent:write` permission.

### Component

**New file:** `ui/src/components/domain/pending-agent-card.tsx`

A card showing:
- Machine ID (full)
- Hostname
- Fingerprint (full SHA256)
- Requested time (relative, e.g., "5m ago")
- Accept button (green)
- Reject button (red/muted)

Uses existing components: `Card`, `Badge variant="pending"`, `Button`,
`Text`.

### Data

Use the generated SDK function `getAgentsPending()` from the agent
operations module. Poll on the same interval as the dashboard health
data.

### Accept/Reject

Call `acceptAgent(hostname)` and `rejectAgent(hostname)` from the
generated SDK. On success, refresh the pending list. Show a brief
success/error state on the button.

## 4. Command Bar

**File:** Register commands in the Dashboard page (`ui/src/pages/dashboard.tsx`)

Commands:
- `pending` — scroll to or expand the pending agents section
- `accept <hostname>` — accept a pending agent by hostname
- `reject <hostname>` — reject a pending agent by hostname

Commands are registered via `useCommands()` hook. Accept/reject
commands need a hostname argument — use the command bar's input
to extract it.

## 5. Components Summary

| Component | File | New/Modify |
|---|---|---|
| `AgentCard` | `agent-card.tsx` | Modify |
| `PendingAgentCard` | `pending-agent-card.tsx` | New |
| `statusBadgeVariant` | `jobs.tsx` | Modify |
| Dashboard | `dashboard.tsx` | Modify |
| Command registration | `dashboard.tsx` | Modify |

## Non-Goals

- Full agent detail page (separate feature)
- PKI configuration UI (config is YAML-only)
- Key rotation UI (CLI-only for now)
