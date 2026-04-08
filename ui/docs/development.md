# Development

This guide covers the tools, setup, and conventions needed to work on
osapi-ui.

## Prerequisites

Install tools using [mise][]:

```bash
mise install
```

- **[Node.js][]** — Required runtime.
- **[Bun][]** — Package manager and script runner.
- **[just][]** — Task runner used for building, testing, formatting, and other
  development workflows. Install with `brew install just`.

### Claude Code

If you use [Claude Code][] for development, install these plugins from the
default marketplace:

```
/plugin install commit-commands@claude-plugins-official
/plugin install superpowers@claude-plugins-official
```

- **commit-commands** — provides `/commit` and `/commit-push-pr` slash commands
  that follow the project's commit conventions automatically.
- **superpowers** — provides structured workflows for planning, TDD, debugging,
  code review, and git worktree isolation.

## Setup

Fetch shared justfiles and install all dependencies:

```bash
just fetch
just deps
```

## Development server

```bash
just dev
```

Opens at `http://localhost:5173`. Hot-reloads on file changes.

## Environment variables

Create a `.env.local` file (gitignored):

```bash
OSAPI_API_URL=http://localhost:8080
OSAPI_BEARER_TOKEN=<your-jwt-here>
```

All env vars use the `OSAPI_` prefix (configured in `vite.config.ts`).

| Variable               | Default                | Description              |
| ---------------------- | ---------------------- | ------------------------ |
| `OSAPI_API_URL`        | `http://localhost:8080` | OSAPI API base URL      |
| `OSAPI_BEARER_TOKEN`   | (empty)                | JWT token for auto-login |
| `OSAPI_FEATURE_STACKS` | `false`                | Enable saved stacks UI   |

Generate a bearer token with the OSAPI CLI:

```bash
osapi token generate
```

If `OSAPI_BEARER_TOKEN` is set, the app auto-authenticates and skips the
sign-in page. If not set, users paste their token on the sign-in page.

## Code style

TypeScript and CSS should be formatted by [Prettier][] and linted using
[ESLint][]. This style is enforced by CI.

```bash
just react::fmt           # Auto-fix formatting
just react::lint          # Run ESLint
```

### Component conventions

- One component per file.
- Use `cva` from class-variance-authority for component variants.
- Use the `cn()` helper for conditional Tailwind classes.
- Icons from lucide-react only.
- No inline styles — Tailwind classes only.
- Use shared UI primitives (see [Architecture](architecture.md) docs) instead
  of repeating Tailwind patterns inline.
- **Always use `Text` for styled text** — never write `text-xs text-text-muted`
  inline. Use `<Text variant="muted">` instead.
- Always use the custom `Dropdown` component. Never use native `<select>`.
- Use Tailwind scale only (`text-xs`, `text-sm`, etc.). Never use arbitrary
  pixel values like `text-[10px]`.

### File naming

- Components: `kebab-case.tsx` (e.g., `agent-card.tsx`)
- Hooks: `use-kebab-case.ts` (e.g., `use-health.ts`)
- Utilities: `kebab-case.ts` (e.g., `cn.ts`)

## SDK regeneration

When the OSAPI Go API changes, copy the combined OpenAPI spec from the osapi
repo and regenerate the TypeScript SDK:

```bash
# In the osapi repo, run `just generate` first to produce the combined spec.
# Then copy it into this repo:
cp <osapi-repo>/internal/controller/api/gen/api.yaml src/sdk/gen/api.yaml

# Regenerate typed fetch functions from the spec:
just react::generate
```

See the [Architecture](architecture.md) docs for the full generation flow.

## Before committing

Run `just ready` before committing to ensure SDK generation, formatting, lint,
and build are all up to date:

```bash
just ready   # generate + fmt + lint + build
```

## Branching

All changes should be developed on feature branches. Create a branch from
`main` using the naming convention `type/short-description`, where `type`
matches the [Conventional Commits][] type:

- `feat/add-dns-block`
- `fix/agent-card-overflow`
- `docs/update-architecture`
- `refactor/extract-component`
- `chore/update-dependencies`

When using Claude Code's `/commit` command, a branch will be created
automatically if you are on `main`.

## Commit messages

Follow [Conventional Commits][] with the 50/72 rule:

- **Subject line**: max 50 characters, imperative mood, capitalized, no period
- **Body**: wrap at 72 characters, separated from subject by a blank line
- **Format**: `type(scope): description`
- **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`,
  `chore`
- Summarize the "what" and "why", not the "how"

Try to write meaningful commit messages and avoid having too many commits on a
PR. Most PRs should likely have a single commit (although for bigger PRs it may
be reasonable to split it in a few). Git squash and rebase is your friend!

[mise]: https://mise.jdx.dev
[Node.js]: https://nodejs.org
[Bun]: https://bun.sh
[just]: https://just.systems
[Claude Code]: https://claude.ai/code
[Prettier]: https://prettier.io
[ESLint]: https://eslint.org
[Conventional Commits]: https://www.conventionalcommits.org
