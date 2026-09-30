---
sidebar_position: 6
---

# UI Architecture

OSAPI ships with an embedded management dashboard: a React single-page
application served from the same host and port as the REST API, giving operators
fleet health, agents, jobs, and block-based operation composition without a
separate frontend to deploy. This page is what an operator acts on — the setting
that turns the UI off, the roles that decide what it shows, and the pages
themselves. How it is built, embedded, structured and generated is stated in the
[specifications repository](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/007-the-embedded-ui/spec.md),
not here.

## Configuration

The UI can be disabled by setting `controller.ui.enabled: false` in
`osapi.yaml`. When disabled, the controller skips registering the SPA handler
and serves only the REST API.

```yaml
controller:
  ui:
    enabled: true # default: true
```

The UI is served from the same host and port as the REST API
(`controller.api.port`), so there is no additional network configuration.

## Authentication & Authorization

The UI uses the same JWT-based auth as the rest of OSAPI. Tokens are generated
via `osapi token generate` and contain a `roles` claim with an array of role
strings (`admin`, `write`, `read`).

### RBAC model

Three built-in roles with hierarchical permissions:

| Role     | JWT value | Permissions                            |
| -------- | --------- | -------------------------------------- |
| Admin    | `admin`   | All permissions including `audit:read` |
| Operator | `write`   | Read + write + execute (no audit)      |
| Viewer   | `read`    | Read-only access                       |

Permissions use `resource:verb` format matching osapi's Go model: `agent:read`,
`file:write`, `command:execute`, `docker:execute`, etc. Configure blocks map to
required permissions in `BLOCK_PERMISSIONS` (`ui/src/lib/permissions.ts`);
unauthorized blocks are shown greyed out with a lock icon.

## Pages

### Dashboard (`/`)

Fleet health overview: summary stat cards, controller and NATS server component
health with hostnames and resource usage, JetStream stream and consumer counts,
KV store and object store usage, and agent cards with status, conditions,
labels, and drain/undrain actions.

### Configure (`/configure`)

Block-based operations builder: sidebar with block categories (Cron, File,
Docker, Command, DNS, Network), blocks gated by RBAC permissions, per-block
target picker (`_all`, `_any`, hostname, labels), sequential apply with
per-block spinners, and result rendering.

### Roles (`/roles`)

RBAC reference: current session info with role badge, role definitions table,
full permission matrix, and block permissions table.

### Enrollment (`/admin/enrollment`)

PKI enrollment management: lists pending agents with machine ID, fingerprint,
and requested time. Accept and reject buttons for each agent. Empty state when
no agents are pending. Command bar commands for accept/reject by hostname.

### SignIn

JWT token authentication: token paste field with validation, role extraction
from JWT claims, and a CLI hint for `osapi token generate`.
