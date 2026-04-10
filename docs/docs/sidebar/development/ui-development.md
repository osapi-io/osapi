---
sidebar_position: 4
---

# UI Development

The embedded management dashboard lives in the `ui/` directory. It is a React 19
SPA built with TypeScript, Vite, and Tailwind CSS v4. The production build is
embedded into the Go binary via `//go:embed` and served by the controller API.

## Prerequisites

Install tools using [mise][]:

```bash
mise install
```

- **[Node.js][]** - JavaScript runtime (Node.js 22).
- **[Bun][]** - Package manager and script runner for the UI.

These are configured in `.mise.toml` alongside the Go toolchain, so a single
`mise install` sets up both the backend and frontend environments.

## Setup

Fetch shared justfiles and install all dependencies:

```bash
just fetch
just deps
```

This installs Go modules, Docusaurus dependencies, and the UI's Bun packages. To
install only the UI dependencies:

```bash
just react::deps
```

## Development server

```bash
just react::dev
```

Opens at `http://localhost:5173`. Hot-reloads on file changes.

For standalone UI development against a running osapi instance, create
`ui/.env.local` (gitignored):

```bash
OSAPI_API_URL=http://localhost:8080
OSAPI_BEARER_TOKEN=<your-jwt-here>
```

Generate a bearer token with the OSAPI CLI:

```bash
osapi token generate
```

If `OSAPI_BEARER_TOKEN` is set, the app auto-authenticates and skips the sign-in
page. If not set, users paste their token on the sign-in page.

## Production build

```bash
just build
```

This is the top-level build recipe. It runs `just react::build` first (to
populate `ui/dist/` with static assets), then `just go::build` to produce the Go
binary with the assets embedded via the `//go:embed` directive in `ui/embed.go`.
The controller API serves these assets at runtime from the embedded filesystem
— no separate web server is required.

:::important

`go build` directly (without building the UI first) will fail because
`//go:embed dist/*` requires at least one file in `ui/dist/` at compile time.
Always use `just build`, `just ready`, or `just test` — they all build the UI
before compiling Go code. The same applies to `go test ./...` for the same
reason.

:::

## Environment variables

All UI env vars use the `OSAPI_` prefix (configured in `vite.config.ts`).

| Variable               | Default | Description                        |
| ---------------------- | ------- | ---------------------------------- |
| `OSAPI_API_URL`        | (empty) | API base URL (empty = same origin) |
| `OSAPI_BEARER_TOKEN`   | (empty) | JWT token for auto-login           |
| `OSAPI_FEATURE_STACKS` | `false` | Enable saved stacks UI             |

When the UI is served from the embedded Go binary, `OSAPI_API_URL` is left empty
so API calls resolve to the same origin as the page.

## Code style

TypeScript and CSS are formatted by [Prettier][] and linted using [ESLint][].
This style is enforced by CI.

```bash
just react::fmt     # Auto-fix formatting
just react::lint    # Run ESLint
```

### Component conventions

- One component per file.
- Use `cva` from class-variance-authority for component variants.
- Use the `cn()` helper for conditional Tailwind classes.
- Icons from lucide-react only.
- No inline styles — Tailwind classes only.
- **Always use the `Text` component for styled text** — never write
  `text-xs text-text-muted` inline. Use `<Text variant="muted">` instead.
- **Always use the custom `Dropdown` component.** Never use native `<select>`.
- Use the Tailwind scale only (`text-xs`, `text-sm`, etc.). Never use arbitrary
  pixel values like `text-[10px]`.
- **Every visual pattern must be a component.** Never duplicate styling with raw
  Tailwind when a component exists.

### File naming

- Components: `kebab-case.tsx` (e.g., `agent-card.tsx`)
- Hooks: `use-kebab-case.ts` (e.g., `use-health.ts`)
- Utilities: `kebab-case.ts` (e.g., `cn.ts`)

## SDK regeneration

The UI uses a generated TypeScript SDK under `ui/src/sdk/gen/`. When the Go API
changes, regenerate both the Go and TypeScript SDKs in a single pass from the
repository root:

```bash
just generate
```

This runs `redocly join` to produce the combined OpenAPI spec, `go generate` for
the Go SDK, copies the combined spec to `ui/src/sdk/gen/api.yaml`, and runs
[orval][] to regenerate the TypeScript SDK.

Never edit files under `ui/src/sdk/gen/` manually. The only hand-written file in
`ui/src/sdk/` is `fetch.ts`, which wires auth and the base URL into the
generated fetch functions.

## Before committing

Run `just ready` from the repository root before committing to ensure all code
generation, formatting, lint, and build steps are up to date:

```bash
just ready
```

This runs `just generate`, formatters, linters, and both the Go and UI builds so
your commit matches what CI will verify.

[mise]: https://mise.jdx.dev
[Node.js]: https://nodejs.org
[Bun]: https://bun.sh
[Prettier]: https://prettier.io
[ESLint]: https://eslint.org
[orval]: https://orval.dev
