---
sidebar_position: 6
---

# UI Architecture

OSAPI ships with an embedded management dashboard — a React single-page
application that lives in the `ui/` directory and is compiled into the Go binary
at build time. Operators get a web UI for fleet health, agents, jobs, and
block-based operation composition without having to deploy a separate frontend.

## Embedding Mechanism

The UI is packaged into the Go binary using `//go:embed`. The React build output
and the embed directive both live in the same Go package:

```
ui/
  dist/            — Vite production build output (generated)
  embed.go         — //go:embed dist/* → ui.Assets
  src/             — React source
  package.json
  vite.config.ts
```

`ui/embed.go` declares a single exported symbol:

```go
package ui

import "embed"

//go:embed dist/*
var Assets embed.FS
```

At runtime, the controller API wires this embedded filesystem into an Echo
handler under `internal/controller/api/ui/`. The handler serves static files
directly and falls back to `index.html` for any non-`/api/` path so React Router
can handle client-side routing.

```
osapi binary
  ├─ embedded dist/  (from ui/embed.go)
  └─ controller API
        ├─ /api/*       → domain handlers (agents, jobs, health, …)
        └─ /*           → internal/controller/api/ui (SPA handler)
```

Because the UI is compiled into the binary, deploying the dashboard is as simple
as deploying `osapi` itself. There are no separate static asset servers, nginx
configs, or container images to manage.

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

## Application Structure

```
ui/
  src/
    App.tsx                   — Route definitions, AuthProvider wrapper
    main.tsx                  — Entry point
    index.css                 — Tailwind theme (colors, fonts)

    sdk/
      gen/                    — Generated SDK (orval) — DO NOT EDIT
        api.yaml              — Combined OpenAPI spec (copied from osapi)
        schemas/              — Generated TypeScript types
        {domain}/             — Generated fetch functions per API domain
      fetch.ts                — Custom fetch mutator (auth + base URL)

    lib/
      auth.tsx                — React auth context (JWT, role resolution)
      permissions.ts          — RBAC roles, permissions, block→permission map
      features.ts             — Feature flags from env vars
      cn.ts                   — Tailwind class merging helper

    components/
      ui/                     — Reusable primitives (framework-level)
      layout/                 — Page structure (Navbar, PageLayout, etc.)
      domain/                 — Domain-specific components (blocks, cards, etc.)

    hooks/                    — Data fetching and state management
    pages/                    — Route page components

  embed.go                    — //go:embed dist/* directive
  dist/                       — Vite production build output
```

## Tech Stack

- **React 19** — UI framework
- **TypeScript** — type-safe codebase
- **Vite** — build tool and dev server
- **Tailwind CSS v4** — utility-first styling via CSS-based `@theme`
- **class-variance-authority (cva)** — component variant patterns
- **clsx + tailwind-merge** — conditional class composition (`cn()` helper)
- **lucide-react** — icons
- **React Router v7** — client-side routing
- **orval** — OpenAPI SDK generation

## Component Architecture

The UI is organized in layers. Pages compose domain components, which compose UI
primitives, which consume Tailwind theme tokens:

```
Pages (Dashboard, Configure, Roles, SignIn)
  └─ Domain Components (AgentCard, BlockCard, ResultCard, …)
       └─ UI Primitives (Card, Button, Badge, Input, Dropdown, …)
            └─ Tailwind CSS theme tokens
```

### UI Primitives (`ui/src/components/ui/`)

Framework-level reusable components that define the visual language: `Text`,
`Card`, `Button`, `Badge`, `Input`, `Dropdown`, `FormField`, `PageHeader`,
`SectionLabel`, `StatCard`, `DataTable`, `HealthDot`, `ErrorBanner`, `Modal`,
`EmptyState`, `Popover`, and more. Every visual pattern is a component — raw
Tailwind classes are never duplicated inline.

### Domain Components (`ui/src/components/domain/`)

Business logic components specific to OSAPI: the block system (`BlockCard`,
`BlockStack`, `ResultCard`), per-operation block forms (`CommandBlock`,
`CronBlock`, `DockerBlock`, `FileBlock`, …), pickers (`TargetPicker`,
`ObjectPicker`, `ContainerPicker`), and dashboard widgets (`AgentCard`,
`ComponentRow`, `JobDetail`).

### Layout Components (`ui/src/components/layout/`)

Page-level structure: `Navbar`, `PageLayout`, `ContentArea`, and the
canvas-based `NetworkMapBackground` animation.

### Hooks (`ui/src/hooks/`)

Data fetching and state management: `useHealth`, `useAgents`, `useStack`,
`useStacks`, `useTargets`, `useObjects`, `useFacts`, and keyboard navigation
helpers like `usePopoverKeyboard`.

## Authentication & Authorization

The UI uses the same JWT-based auth as the rest of OSAPI. Tokens are generated
via `osapi token generate` and contain a `roles` claim with an array of role
strings (`admin`, `write`, `read`).

### Auth flow

```
User opens app
  → No token in localStorage?
    → SignIn page → paste JWT from `osapi token generate`
    → Token decoded client-side → roles extracted
    → Stored in localStorage
  → Token exists (or OSAPI_BEARER_TOKEN env var set)?
    → Auto-authenticate → Dashboard
```

The UI decodes the token client-side (no verification — that is the server's
job) to extract roles. The token is sent as a `Bearer` header on every API
request via the fetch mutator in `ui/src/sdk/fetch.ts`.

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

## SDK Generation

The TypeScript SDK is generated from OSAPI's combined OpenAPI spec using
[orval][]. The Go SDK and TypeScript SDK both flow from the same source of
truth.

```
osapi repo
┌─────────────────────────────────────────────┐
│ internal/controller/api/{domain}/gen/       │
│   api.yaml                                  │
│ internal/controller/api/gen/api.yaml        │
│   (combined spec via redocly join)          │
│                                             │
│   ── copy ──►  ui/src/sdk/gen/api.yaml     │
│                 └─ orval → TypeScript SDK  │
└─────────────────────────────────────────────┘
```

Running `just generate` from the repository root performs the full regeneration
in order:

1. `redocly join` combines the per-domain OpenAPI specs into
   `internal/controller/api/gen/api.yaml`.
2. `go generate` regenerates Go server and SDK code.
3. The combined spec is copied to `ui/src/sdk/gen/api.yaml`.
4. `just react::generate` runs orval against the copied spec to regenerate typed
   fetch functions and schema types under `ui/src/sdk/gen/`.

### Fetch mutator

`ui/src/sdk/fetch.ts` is the only hand-written file in `ui/src/sdk/`. It:

- Reads the base URL from `OSAPI_API_URL` (empty = same origin)
- Gets the auth token from the React auth context module
- Sends it as a `Bearer` header
- Wraps responses in `{ data, status, headers }` for orval

Never edit files under `ui/src/sdk/gen/` — they are overwritten on every
generation.

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

### SignIn

JWT token authentication: token paste field with validation, role extraction
from JWT claims, and a CLI hint for `osapi token generate`.

[orval]: https://orval.dev
