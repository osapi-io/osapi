# Architecture

OSAPI UI is a React single-page application that provides a management
dashboard and operations builder for [OSAPI](https://github.com/osapi-io/osapi).

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

## Application Structure

```
src/
  App.tsx              — Route definitions, AuthProvider wrapper
  main.tsx             — Entry point
  index.css            — Tailwind theme (colors, fonts)

  sdk/
    gen/               — Generated SDK (orval) — DO NOT EDIT
      api.yaml         — Combined OpenAPI spec (copied from osapi)
      schemas/         — Generated TypeScript types
      {domain}/        — Generated fetch functions per API domain
    fetch.ts           — Custom fetch mutator (auth + base URL)

  lib/
    auth.tsx           — React auth context (JWT, role resolution)
    permissions.ts     — RBAC roles, permissions, block→permission map
    features.ts        — Feature flags from env vars
    cn.ts              — Tailwind class merging helper

  components/
    ui/                — Reusable primitives (framework-level)
    layout/            — Page structure (Navbar, PageLayout, etc.)
    domain/            — Domain-specific components (blocks, cards, etc.)

  hooks/               — Data fetching and state management
  pages/               — Route page components
```

## Authentication & Authorization

### Auth Flow

```
User opens app
  → No token in localStorage?
    → SignIn page → paste JWT from `osapi token generate`
    → Token decoded client-side → roles extracted
    → Stored in localStorage
  → Token exists (or OSAPI_BEARER_TOKEN env var set)?
    → Auto-authenticate → Dashboard
```

### JWT Token

OSAPI issues JWTs via `osapi token generate`. The token contains:
- `roles` claim — array of role strings (`admin`, `write`, `read`)
- Standard JWT claims (sub, exp, iat)

The UI decodes the token client-side (no verification — that's the
server's job) to extract roles. The token is sent as a Bearer header
on every API request via the fetch mutator.

### RBAC Model

Three built-in roles with hierarchical permissions:

| Role | JWT Value | Permissions |
| --- | --- | --- |
| Admin | `admin` | All 17 permissions including `audit:read` |
| Operator | `write` | Read + write + execute (no audit) |
| Viewer | `read` | Read-only access |

Permissions use `resource:verb` format matching osapi's Go model:
`agent:read`, `file:write`, `command:execute`, `docker:execute`, etc.

### Permission Gating

- **Configure blocks** — each block type maps to a required permission
  in `BLOCK_PERMISSIONS`. Unauthorized blocks show greyed out with a
  lock icon.
- **Agent drain/undrain** — requires `agent:write`.
- **Role override dropdown** — lets users preview what other roles see
  without changing the token.

## SDK Generation

The TypeScript SDK is generated from OSAPI's combined OpenAPI spec using
[orval](https://orval.dev/).

### Generation Flow

```
osapi repo                          osapi-ui repo
┌────────────────────┐              ┌──────────────────────┐
│ api/{domain}/gen/  │              │ src/sdk/gen/         │
│   api.yaml         │  just gen    │   api.yaml (copy)    │
│   ...              │ ──────────►  │   schemas/           │
│ api/gen/api.yaml   │  (redocly)   │   {domain}/          │
│ (combined spec)    │              │   (orval output)     │
└────────────────────┘              └──────────────────────┘
```

1. In the osapi repo, `just generate` runs oapi-codegen for each domain
   then `redocly join` to produce the combined spec at
   `internal/controller/api/gen/api.yaml`.
2. In osapi-ui, `just generate` copies that combined spec into
   `src/sdk/gen/api.yaml` and runs orval to produce typed fetch
   functions and schema types.

### Fetch Mutator

`src/sdk/fetch.ts` is the only hand-written file in `src/sdk/`. It:
- Reads the base URL from `OSAPI_API_URL` env var
- Gets the auth token from the React auth context module
- Sends it as a `Bearer` header
- Wraps responses in `{ data, status, headers }` for orval

## Component Architecture

### Layers

```
Pages (Dashboard, Configure, Roles, SignIn)
  └─ Domain Components (AgentCard, BlockCard, ResultCard, etc.)
       └─ UI Primitives (Card, Button, Badge, Input, Dropdown, etc.)
            └─ Tailwind CSS theme tokens
```

### UI Primitives (`src/components/ui/`)

Framework-level reusable components. These define the visual language:

| Component | Purpose |
| --- | --- |
| Text | Styled text with variant/size props — the default for all text |
| Card, CardHeader, CardTitle, CardContent | Container with variant borders/shadows |
| Button | Primary/secondary/ghost/destructive with sizes |
| Badge | Status indicators (ready/pending/running/error/applied/muted) |
| Input | Form input with label, autofill suppression |
| Dropdown | Custom popover dropdown (replaces native select) |
| FactInput | Input with @fact. reference autocomplete |
| FormField | Label + Input wrapper for consistent form layout |
| PageHeader | Page title + subtitle + optional actions slot |
| SectionLabel | Uppercase section header with optional icon |
| StatCard | Label + big value + detail text in a card |
| DataTable | Typed table with header/rows inside a card |
| HealthDot | Colored status dot (ok/error/muted) |
| ErrorBanner | Error message with icon (sm/md sizes) |
| MetricValue | Formatted metric display (label + value) |
| CodeBlock | Styled code/pre block with border |
| IdBadge | Monospace ID pill for identifiers |
| IconButton | Icon-only button with ghost/danger/accent variants |
| CollapsibleSection | Togglable section with chevron, icon, right content |
| Modal | Dialog overlay with close button |
| EmptyState | Centered message with dashed border and optional icon |
| SearchBox | Inline search input with close button |
| ScrollButton | Directional scroll arrow (left/right) |
| InfoBox | Subtle container for hints and info text |
| KeyValue | Inline key:value display pair |
| LabelTag | Accent-colored key:value label pill |
| ConditionAlert | Warning condition with triangle icon |
| Popover, PopoverItem, PopoverPanel | Floating popover menu system |

### Text Component

The `Text` component is the standard way to render styled text. Never
write inline Tailwind text classes — always use `Text` with the
appropriate variant:

```tsx
<Text variant="muted">Secondary text</Text>
<Text variant="mono-muted">code_value</Text>
<Text variant="error" as="p">Error message</Text>
<Text size="sm" className="font-medium">Title text</Text>
<Text variant="accent">@fact.reference</Text>
```

Available variants: `default`, `muted`, `label`, `mono`, `mono-muted`,
`mono-primary`, `error`, `primary`, `accent`.

Available sizes: `xs` (default), `sm`, `base`.

### Domain Components (`src/components/domain/`)

Business logic components specific to OSAPI:

- **Block system** — BlockCard, BlockStack, ApplyButton, ResultCard,
  SaveStackDialog, StackBar
- **Block forms** — one per operation type (CommandBlock, CronBlock,
  DockerBlock, DockerExecBlock, FileBlock, FileUploadBlock,
  FileDeleteBlock, CronDeleteBlock, ContainerActionBlock,
  DnsUpdateBlock, SingleInputBlock)
- **Pickers** — TargetPicker (agents/labels), ObjectPicker (files),
  ContainerPicker, CronPicker
- **Dashboard** — AgentCard, ComponentRow, JobDetail, HostGroupHeader

### Hooks (`src/hooks/`)

| Hook | Purpose |
| --- | --- |
| useHealth | Poll `/health/status` every 10s |
| useAgents | Poll `/agent` every 10s with refresh callback |
| useStack | Block state management (add, remove, apply, reset) |
| useStacks | Saved stack management (behind feature flag) |
| useTargets | Build target options from agent list |
| useObjects | Fetch file objects for pickers |
| useFacts | Fetch @fact. keys from `/facts/keys` |

## Feature Flags

Feature flags gate unreleased functionality via env vars:

| Flag | Default | Description |
| --- | --- | --- |
| `OSAPI_FEATURE_STACKS` | `false` | Saved stacks UI (pending Go API) |

## Pages

### Dashboard (`/`)

Fleet health overview:
- Summary stat cards (NATS, Jobs, Agents, Consumers)
- Controller and NATS Server component health with hostname/cpu/memory
- JetStream streams table with message counts and consumer counts
- KV stores table with key counts and sizes
- Object store table with sizes
- Agent cards with status, conditions, labels, drain/undrain

### Configure (`/configure`)

Block-based operations builder:
- Sidebar with block categories (Cron, File, Docker, Command, DNS, Network)
- Blocks gated by RBAC permissions
- Per-block target picker (\_all, \_any, hostname, labels)
- Sequential apply with per-block spinners and result rendering
- Saved stacks bar (behind feature flag)

### Roles (`/roles`)

RBAC reference:
- Current session info with role badge
- Role definitions table
- Full permission matrix (permission × role)
- Block permissions table (block type → required permission × role)

### SignIn

JWT token authentication:
- Token paste field with validation
- Role extraction from JWT claims
- CLI hint for `osapi token generate`
