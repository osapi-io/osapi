# SDK Monorepo Migration (PR 1: Client) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Move the SDK client library from `osapi-io/osapi-sdk` into this repo
as `pkg/sdk/osapi/`, flatten examples, add Docusaurus SDK docs, and update all
references.

**Architecture:** Copy `osapi-sdk/pkg/osapi/` into `pkg/sdk/osapi/`, rewrite
the codegen to read the server's combined spec directly (no gilt), update all 18
Go import paths, flatten 9 example directories into individual files, create
Docusaurus SDK sidebar pages, and clean up CLAUDE.md/README.md references.

**Tech Stack:** Go, oapi-codegen, Docusaurus, Cobra CLI

**Design doc:** `docs/plans/2026-03-07-sdk-monorepo-migration-design.md`

---

### Task 1: Copy SDK client package

**Files:**

- Create: `pkg/sdk/osapi/` (all `.go` files from `osapi-sdk/pkg/osapi/`)
- Create: `pkg/sdk/osapi/gen/` (cfg.yaml, generate.go, client.gen.go)

**Step 1: Copy source files**

```bash
mkdir -p pkg/sdk/osapi/gen
cp ../osapi-sdk/pkg/osapi/*.go pkg/sdk/osapi/
cp ../osapi-sdk/pkg/osapi/gen/client.gen.go pkg/sdk/osapi/gen/
```

**Step 2: Create new cfg.yaml**

Create `pkg/sdk/osapi/gen/cfg.yaml`:

```yaml
---
package: gen
output: client.gen.go
generate:
  models: true
  client: true
output-options:
  skip-prune: true
```

**Step 3: Create new generate.go**

Create `pkg/sdk/osapi/gen/generate.go`:

```go
// Package gen contains generated code for the OSAPI REST API client.
package gen

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml ../../../internal/api/gen/api.yaml
```

No gilt — oapi-codegen reads the server's combined spec directly.

**Step 4: Update package import paths in all copied files**

In every `.go` file under `pkg/sdk/osapi/` (non-test, non-gen), replace:

```
"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
```

with:

```
"github.com/retr0h/osapi/pkg/sdk/osapi/gen"
```

In every `_test.go` file under `pkg/sdk/osapi/`, replace:

```
"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
```

with:

```
"github.com/retr0h/osapi/pkg/sdk/osapi/gen"
```

And for public test files, replace:

```
"github.com/osapi-io/osapi-sdk/pkg/osapi"
```

with:

```
"github.com/retr0h/osapi/pkg/sdk/osapi"
```

**Step 5: Regenerate client to verify**

```bash
cd pkg/sdk/osapi/gen && go generate ./...
```

Verify `client.gen.go` is regenerated without errors.

**Step 6: Commit**

```bash
git add pkg/sdk/
git commit -m "feat(sdk): copy client library into pkg/sdk/osapi"
```

---

### Task 2: Update Go imports and remove external SDK dependency

**Files:**

- Modify: `go.mod` (remove `github.com/osapi-io/osapi-sdk` require)
- Modify: 18 Go files (update import paths)

**Step 1: Update all Go imports**

In every file listed below, replace
`"github.com/osapi-io/osapi-sdk/pkg/osapi"` with
`"github.com/retr0h/osapi/pkg/sdk/osapi"`:

- `cmd/client.go`
- `cmd/client_agent_get.go`
- `cmd/client_audit_export.go`
- `cmd/client_file_upload.go` (aliased import: `osapi "..."`)
- `cmd/client_health_status.go`
- `cmd/client_job_list.go`
- `cmd/client_job_run.go`
- `cmd/client_node_command_exec.go`
- `cmd/client_node_command_shell.go`
- `cmd/client_node_file_deploy.go`
- `cmd/client_node_status_get.go`
- `internal/audit/export/types.go`
- `internal/audit/export/file.go`
- `internal/audit/export/file_test.go`
- `internal/audit/export/export_public_test.go`
- `internal/audit/export/file_public_test.go`
- `internal/cli/ui.go`
- `internal/cli/ui_public_test.go`

**Step 2: Remove external SDK from go.mod**

Remove the `github.com/osapi-io/osapi-sdk` line from the `require` block in
`go.mod`. Then run:

```bash
go mod tidy
```

This will remove the SDK from `go.sum` as well.

**Step 3: Build and test**

```bash
go build ./...
go test ./... -count=1 -timeout 120s
```

**Step 4: Commit**

```bash
git add -A
git commit -m "refactor(sdk): update imports to pkg/sdk/osapi"
```

---

### Task 3: Flatten SDK client examples

**Files:**

- Create: `examples/sdk/osapi/go.mod`
- Create: `examples/sdk/osapi/health.go` (and 8 more)

**Step 1: Create examples directory and go.mod**

```bash
mkdir -p examples/sdk/osapi
```

Create `examples/sdk/osapi/go.mod`:

```
module github.com/retr0h/osapi/examples/sdk/osapi

go 1.25.0

replace github.com/retr0h/osapi => ../../../

require github.com/retr0h/osapi v0.0.0
```

Then run:

```bash
cd examples/sdk/osapi && go mod tidy
```

**Step 2: Create flattened example files**

Copy each example's `main.go` into a single file, updating the import path.
Each file is `package main` and self-contained.

From `../osapi-sdk/examples/osapi/`:

| Source directory | Target file |
|---|---|
| `health/main.go` | `examples/sdk/osapi/health.go` |
| `node/main.go` | `examples/sdk/osapi/node.go` |
| `agent/main.go` | `examples/sdk/osapi/agent.go` |
| `audit/main.go` | `examples/sdk/osapi/audit.go` |
| `command/main.go` | `examples/sdk/osapi/command.go` |
| `file/main.go` | `examples/sdk/osapi/file.go` |
| `job/main.go` | `examples/sdk/osapi/job.go` |
| `metrics/main.go` | `examples/sdk/osapi/metrics.go` |
| `network/main.go` | `examples/sdk/osapi/network.go` |

In each file, replace:

```
"github.com/osapi-io/osapi-sdk/pkg/osapi"
```

with:

```
"github.com/retr0h/osapi/pkg/sdk/osapi"
```

**Step 3: Verify examples compile**

```bash
cd examples/sdk/osapi && go build ./...
```

Note: `go build ./...` on `package main` files in the same directory will
verify they all compile. They won't run without a live server, but compilation
proves imports are correct.

**Step 4: Commit**

```bash
git add examples/sdk/
git commit -m "feat(sdk): add flattened client examples"
```

---

### Task 4: Create Docusaurus SDK client pages

**Files:**

- Create: `docs/docs/sidebar/sdk/sdk.md`
- Create: `docs/docs/sidebar/sdk/client/client.md`
- Create: `docs/docs/sidebar/sdk/client/agent.md`
- Create: `docs/docs/sidebar/sdk/client/audit.md`
- Create: `docs/docs/sidebar/sdk/client/file.md`
- Create: `docs/docs/sidebar/sdk/client/health.md`
- Create: `docs/docs/sidebar/sdk/client/job.md`
- Create: `docs/docs/sidebar/sdk/client/metrics.md`
- Create: `docs/docs/sidebar/sdk/client/node.md`
- Modify: `docs/docusaurus.config.ts`

**Step 1: Create SDK landing page**

Create `docs/docs/sidebar/sdk/sdk.md`:

```markdown
---
sidebar_position: 6
---

# SDK

OSAPI provides a Go SDK for programmatic access to the REST API. The SDK
includes a typed client library and a DAG-based orchestrator for composing
multi-step operations.

<DocCardList />
```

**Step 2: Create client overview page**

Create `docs/docs/sidebar/sdk/client/client.md`. Migrate content from
`osapi-sdk/docs/osapi/README.md`: services table, client options, targeting
table. Adapt to Docusaurus format with `<DocCardList />` for per-service pages.

**Step 3: Create per-service pages**

Create one page per service (`agent.md`, `audit.md`, `file.md`, `health.md`,
`job.md`, `metrics.md`, `node.md`). Migrate content from
`osapi-sdk/docs/osapi/{service}.md`. Each page covers the service methods,
parameters, return types, and a usage example.

**Step 4: Update docusaurus.config.ts**

Add "SDK" to the Features navbar dropdown:

```typescript
{
  label: 'SDK',
  to: 'sidebar/sdk/sdk',
},
```

Update the `specPath` for the API docs plugin from
`../../osapi-sdk/pkg/osapi/gen/api.yaml` to
`../internal/api/gen/api.yaml`.

**Step 5: Update the API docs specPath**

In `docs/docusaurus.config.ts`, change:

```typescript
specPath: '../../osapi-sdk/pkg/osapi/gen/api.yaml',
```

to:

```typescript
specPath: '../internal/api/gen/api.yaml',
```

And remove the GitHub download URL reference to the SDK repo.

**Step 6: Verify docs build**

```bash
cd docs && bun run build
```

**Step 7: Commit**

```bash
git add docs/
git commit -m "docs(sdk): add client library pages to Docusaurus"
```

---

### Task 5: Update CLAUDE.md

**Files:**

- Modify: `CLAUDE.md`

**Step 1: Update architecture section**

Change line ~41 from:

```
- **`osapi-sdk`** - External SDK for programmatic REST API access (sibling repo, linked via `replace` in `go.mod`)
```

to:

```
- **`pkg/sdk/`** - Go SDK for programmatic REST API access (`osapi/` client library, `orchestrator/` DAG runner)
```

**Step 2: Rewrite "Update SDK" Step 5**

Replace the entire Step 5 section (lines ~174-196) with:

```markdown
### Step 5: Update SDK

The SDK client library lives in `pkg/sdk/osapi/`. Its generated HTTP client
uses the same combined OpenAPI spec as the server
(`internal/api/gen/api.yaml`).

**When modifying existing API specs:**

1. Make changes to `internal/api/{domain}/gen/api.yaml` in this repo
2. Run `just generate` to regenerate server code (this also regenerates the
   combined spec via `redocly join`)
3. Run `go generate ./pkg/sdk/osapi/gen/...` to regenerate the SDK client
4. Update the SDK service wrappers in `pkg/sdk/osapi/{domain}.go` if new
   response codes were added
5. Update CLI switch blocks in `cmd/` if new response codes were added

**When adding a new API domain:**

1. Add a service wrapper in `pkg/sdk/osapi/{domain}.go`
2. Run `go generate ./pkg/sdk/osapi/gen/...` to pick up the new domain's
   spec from the combined `api.yaml`
```

**Step 3: Remove sibling repo references**

Remove any remaining references to `osapi-sdk` as a "sibling repo" or
"external" dependency. Keep documentation about the SDK but update paths to
`pkg/sdk/`.

**Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for in-repo SDK"
```

---

### Task 6: Update README.md and system-architecture.md

**Files:**

- Modify: `README.md`
- Modify: `docs/docs/sidebar/architecture/system-architecture.md`

**Step 1: Update README.md**

Replace the "Sister Projects" section. Remove `osapi-sdk` from the sister
projects table (it's now in-repo). Add an SDK link in the Documentation
section:

```markdown
## 📖 Documentation

- [Getting Started](https://osapi-io.github.io/osapi/)
- [Features](https://osapi-io.github.io/osapi/sidebar/features/)
- [SDK](https://osapi-io.github.io/osapi/sidebar/sdk/sdk)
- [CLI Reference](https://osapi-io.github.io/osapi/sidebar/usage/)
- [Architecture](https://osapi-io.github.io/osapi/sidebar/architecture/)
```

If `osapi-orchestrator` is still a separate repo, keep it in sister projects
but remove `osapi-sdk`.

**Step 2: Update system-architecture.md**

Change line ~19 from:

```
| **SDK Client** | `osapi-sdk` (external) | OpenAPI-generated client used by CLI |
```

to:

```
| **SDK Client** | `pkg/sdk/osapi` | OpenAPI-generated client used by CLI |
```

Update the mermaid diagram reference from `SDK["SDK Client (osapi-sdk)"]` to
`SDK["SDK Client (pkg/sdk/osapi)"]`.

**Step 3: Commit**

```bash
git add README.md docs/docs/sidebar/architecture/system-architecture.md
git commit -m "docs: update README and architecture for in-repo SDK"
```

---

### Task 7: Final verification

**Step 1: Full build**

```bash
go build ./...
```

**Step 2: Full test suite**

```bash
go test ./... -count=1 -timeout 120s
```

**Step 3: Lint**

```bash
just go::vet
```

**Step 4: Regenerate to verify codegen pipeline**

```bash
go generate ./pkg/sdk/osapi/gen/...
go build ./...
```

**Step 5: Verify examples compile**

```bash
cd examples/sdk/osapi && go build ./...
```

**Step 6: Verify docs build**

```bash
cd docs && bun run build
```

---

## Files Summary

| Action | Path |
|---|---|
| Create | `pkg/sdk/osapi/*.go` (all source + test files) |
| Create | `pkg/sdk/osapi/gen/cfg.yaml` |
| Create | `pkg/sdk/osapi/gen/generate.go` |
| Create | `pkg/sdk/osapi/gen/client.gen.go` |
| Create | `examples/sdk/osapi/go.mod` |
| Create | `examples/sdk/osapi/*.go` (9 example files) |
| Create | `docs/docs/sidebar/sdk/sdk.md` |
| Create | `docs/docs/sidebar/sdk/client/client.md` |
| Create | `docs/docs/sidebar/sdk/client/{service}.md` (7 files) |
| Modify | `cmd/*.go` (11 files — import path) |
| Modify | `internal/audit/export/*.go` (5 files — import path) |
| Modify | `internal/cli/ui.go`, `ui_public_test.go` (import path) |
| Modify | `go.mod` (remove external SDK) |
| Modify | `CLAUDE.md` |
| Modify | `README.md` |
| Modify | `docs/docs/sidebar/architecture/system-architecture.md` |
| Modify | `docs/docusaurus.config.ts` |
