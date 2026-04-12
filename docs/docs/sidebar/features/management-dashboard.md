---
sidebar_position: 21
---

# Management Dashboard

OSAPI includes an embedded React management dashboard served directly from the
controller binary. No separate web server, container image, or static file
hosting is required — deploy the `osapi` binary and the UI is ready.

## Overview

The dashboard provides a visual interface for fleet management:

- **Dashboard** — fleet health overview with agent cards (machine ID,
  fingerprint, scheduling state), component status, JetStream metrics, and node
  conditions
- **Configure** — block-based operations builder for composing and applying
  changes across targets
- **Admin** — audit log viewer with export, job queue browser with retry/delete,
  PKI enrollment management (accept/reject agents), and RBAC reference
  (admin-only)

## Enabling the UI

The dashboard is enabled by default. Disable it with:

```yaml
controller:
  ui:
    enabled: false
```

When disabled, the controller serves only the REST API.

## Authentication

The UI uses the same JWT-based authentication as the CLI and API. Generate a
token with `osapi token generate` and paste it on the sign-in page. The token's
`roles` claim determines what the user can see and do:

| Role     | Access                                             |
| -------- | -------------------------------------------------- |
| Admin    | Full access including Audit, Jobs, and role switch |
| Operator | Dashboard + Configure (block operations)           |
| Viewer   | Dashboard only (read-only)                         |

## Architecture

The UI is a React 19 SPA built with Vite and embedded into the Go binary via
`//go:embed`. The controller serves static assets at `/` and falls back to
`index.html` for client-side routing. All API endpoints are prefixed with
`/api/`.

See [UI Architecture](../architecture/ui.md) for details on the embedding
mechanism, component layers, and SDK generation flow.

## Development

See [UI Development](../development/ui-development.md) for prerequisites, the
dev server, code style, and component conventions.
