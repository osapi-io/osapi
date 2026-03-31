# User & Group Management Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add user and group management as a node provider with CRUD operations, password changes, and full API/CLI/SDK support.

**Architecture:** Direct provider at `provider/node/user/` using `useradd`, `usermod`, `userdel`, `groupadd`, `groupmod`, `groupdel`, and `chpasswd` via `exec.Manager`. Parses `/etc/passwd` and `/etc/group` for reads. Two API path prefixes (`user/` and `group/`), two SDK services (`UserService` and `GroupService`). Permissions: `user:read` (all roles), `user:write` (admin + write).

**Tech Stack:** Go 1.25, Echo, oapi-codegen (strict-server), gomock, testify/suite, avfs

**Coverage baseline:** 99.9% — must remain at or above this.

---

## Task 1: SDK Constants (Operations + Permissions)

**Files:**
- Modify: `pkg/sdk/client/operations.go`
- Modify: `pkg/sdk/client/permissions.go`
- Modify: `internal/job/types.go`
- Modify: `internal/authtoken/permissions.go`

- [ ] **Step 1: Add user and group operation constants**

In `pkg/sdk/client/operations.go`:

```go
// User operations.
const (
	OpUserList           JobOperation = "node.user.list"
	OpUserGet            JobOperation = "node.user.get"
	OpUserCreate         JobOperation = "node.user.create"
	OpUserUpdate         JobOperation = "node.user.update"
	OpUserDelete         JobOperation = "node.user.delete"
	OpUserChangePassword JobOperation = "node.user.password"
)

// Group operations.
const (
	OpGroupList   JobOperation = "node.group.list"
	OpGroupGet    JobOperation = "node.group.get"
	OpGroupCreate JobOperation = "node.group.create"
	OpGroupUpdate JobOperation = "node.group.update"
	OpGroupDelete JobOperation = "node.group.delete"
)
```

- [ ] **Step 2: Add permission constants**

```go
	PermUserRead  Permission = "user:read"
	PermUserWrite Permission = "user:write"
```

- [ ] **Step 3: Re-export in internal/job/types.go**

- [ ] **Step 4: Re-export permissions in internal/authtoken/permissions.go**

Add to `DefaultRolePermissions`:
- `RoleAdmin`: `PermUserRead` + `PermUserWrite`
- `RoleWrite`: `PermUserRead` + `PermUserWrite`
- `RoleRead`: `PermUserRead` only

- [ ] **Step 5: Verify and commit**

```bash
go build ./...
git commit -m "feat(user): add operation and permission constants"
```

---

## Task 2: Provider Interface + Platform Stubs

**Files:**
- Create: `internal/provider/node/user/types.go`
- Create: `internal/provider/node/user/darwin.go`
- Create: `internal/provider/node/user/linux.go`
- Create: `internal/provider/node/user/mocks/generate.go`

- [ ] **Step 1: Create types.go**

Provider interface with 11 methods (6 user + 5 group). All data
types: User, Group, CreateUserOpts, UpdateUserOpts,
CreateGroupOpts, UpdateGroupOpts, UserResult, GroupResult.
See design spec for exact type definitions.

- [ ] **Step 2: Create darwin.go and linux.go stubs**

All 11 methods return `fmt.Errorf("user: %w", provider.ErrUnsupported)`.

- [ ] **Step 3: Create mocks and generate**

- [ ] **Step 4: Verify and commit**

```bash
go generate ./internal/provider/node/user/mocks/...
go build ./...
git commit -m "feat(user): add provider interface and platform stubs"
```

---

## Task 3: Debian Provider — User Operations

**Files:**
- Create: `internal/provider/node/user/debian.go`
- Create: `internal/provider/node/user/debian_user.go`
- Create: `internal/provider/node/user/debian_public_test.go`
- Create: `internal/provider/node/user/darwin_public_test.go`
- Create: `internal/provider/node/user/linux_public_test.go`

Split the Debian implementation across two files for readability:
`debian.go` for the struct/constructor + group methods,
`debian_user.go` for user methods. Or organize by concern:
`debian.go` for struct/constructor, `debian_user.go` for user ops,
`debian_group.go` for group ops.

- [ ] **Step 1: Write stub tests (Darwin, Linux)**

Verify all 11 methods return `ErrUnsupported`.

- [ ] **Step 2: Write Debian user tests**

Test cases for each user method:

**TestListUsers:**
- success (parse /etc/passwd, filter UID >= 1000)
- parse error
- empty result (no non-system users)

**TestGetUser:**
- success (user exists)
- user not found
- exec error (id -Gn fails)

**TestCreateUser:**
- success with minimal opts (name only)
- success with all opts (UID, home, shell, groups, password, system)
- useradd error (user already exists)
- password set after creation

**TestUpdateUser:**
- success changing shell
- success changing groups
- success locking user
- success unlocking user
- usermod error

**TestDeleteUser:**
- success
- user not found
- userdel error

**TestChangePassword:**
- success
- chpasswd error

- [ ] **Step 3: Implement debian.go + debian_user.go**

```go
type Debian struct {
	provider.FactsAware
	logger      *slog.Logger
	fs          avfs.VFS
	execManager exec.Manager
}

func NewDebianProvider(
	logger *slog.Logger,
	fs avfs.VFS,
	execManager exec.Manager,
) *Debian
```

Use `avfs.VFS` for reading `/etc/passwd` and `/etc/group` (testable
with memfs). Use `exec.Manager` for running useradd/usermod/etc.

Parsing `/etc/passwd`: each line is `name:x:uid:gid:gecos:home:shell`.
Filter UID >= 1000 for ListUsers. GetUser reads all UIDs.

For ChangePassword: construct `name:password` string and pipe to
`chpasswd`. Read exec.Manager to understand how to pass stdin.
If exec.Manager doesn't support stdin, use
`exec.Manager.RunCmd("chpasswd", []string{})` with the input as
a separate call, or use `usermod --password $(openssl passwd -6 password)`.

- [ ] **Step 4: Verify user tests pass with 100% coverage**

---

## Task 4: Debian Provider — Group Operations

**Files:**
- Create: `internal/provider/node/user/debian_group.go`
- Modify: `internal/provider/node/user/debian_public_test.go` (add group tests)

- [ ] **Step 1: Write Debian group tests**

**TestListGroups:**
- success (parse /etc/group)
- parse error

**TestGetGroup:**
- success
- group not found

**TestCreateGroup:**
- success with name only
- success with GID and system flag
- groupadd error

**TestUpdateGroup:**
- success updating members
- gpasswd error

**TestDeleteGroup:**
- success
- groupdel error

- [ ] **Step 2: Implement debian_group.go**

Parsing `/etc/group`: each line is `name:x:gid:member1,member2`.

- [ ] **Step 3: Verify all tests pass with 100% coverage**

```bash
go test -coverprofile=/tmp/c.out ./internal/provider/node/user/...
git commit -m "feat(user): implement Debian user and group provider with tests"
```

---

## Task 5: Agent Processor + Wiring

**Files:**
- Create: `internal/agent/processor_user.go`
- Create: `internal/agent/processor_user_public_test.go`
- Modify: `internal/agent/processor.go`
- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Create processor with tests**

Two base operations dispatching to sub-operations:

`user.*` operations:
- `user.list` — no data, call `provider.ListUsers(ctx)`
- `user.get` — unmarshal `{"name": "..."}`, call `provider.GetUser`
- `user.create` — unmarshal `CreateUserOpts`, call `provider.CreateUser`
- `user.update` — unmarshal `{"name": "...", ...opts}`, call `provider.UpdateUser`
- `user.delete` — unmarshal `{"name": "..."}`, call `provider.DeleteUser`
- `user.password` — unmarshal `{"name": "...", "password": "..."}`, call `provider.ChangePassword`

`group.*` operations:
- `group.list` — no data, call `provider.ListGroups(ctx)`
- `group.get` — unmarshal `{"name": "..."}`, call `provider.GetGroup`
- `group.create` — unmarshal `CreateGroupOpts`, call `provider.CreateGroup`
- `group.update` — unmarshal `{"name": "...", ...opts}`, call `provider.UpdateGroup`
- `group.delete` — unmarshal `{"name": "..."}`, call `provider.DeleteGroup`

- [ ] **Step 2: Add to node processor**

Add `userProvider user.Provider` to `NewNodeProcessor`. Add
`case "user":` and `case "group":` dispatch — both route to the
same provider but different methods.

- [ ] **Step 3: Wire in agent_setup.go**

```go
func createUserProvider(
	log *slog.Logger,
	fs avfs.VFS,
	execManager exec.Manager,
) userProv.Provider
```

Container check: return `ErrUnsupported` in containers.

- [ ] **Step 4: Fix existing tests and verify**

```bash
go test ./internal/agent/... ./cmd/...
git commit -m "feat(user): add agent processor and wiring"
```

---

## Task 6: OpenAPI Spec + User API Handlers

**Files:**
- Create: `internal/controller/api/node/user/gen/api.yaml`
- Create: `internal/controller/api/node/user/gen/cfg.yaml`
- Create: `internal/controller/api/node/user/gen/generate.go`
- Create: `internal/controller/api/node/user/types.go`
- Create: `internal/controller/api/node/user/user.go`
- Create: `internal/controller/api/node/user/validate.go`
- Create: `internal/controller/api/node/user/user_list_get.go`
- Create: `internal/controller/api/node/user/user_get.go`
- Create: `internal/controller/api/node/user/user_create.go`
- Create: `internal/controller/api/node/user/user_update.go`
- Create: `internal/controller/api/node/user/user_delete.go`
- Create: `internal/controller/api/node/user/user_password.go`
- Create: `internal/controller/api/node/user/handler.go`
- Create: test files for each handler
- Modify: `cmd/controller_setup.go`

- [ ] **Step 1: Create OpenAPI spec**

The spec includes BOTH user and group endpoints. Six user paths:
- `GET /node/{hostname}/user` — list users
- `POST /node/{hostname}/user` — create user
- `GET /node/{hostname}/user/{name}` — get user
- `PUT /node/{hostname}/user/{name}` — update user
- `DELETE /node/{hostname}/user/{name}` — delete user
- `POST /node/{hostname}/user/{name}/password` — change password

Five group paths:
- `GET /node/{hostname}/group` — list groups
- `POST /node/{hostname}/group` — create group
- `GET /node/{hostname}/group/{name}` — get group
- `PUT /node/{hostname}/group/{name}` — update group
- `DELETE /node/{hostname}/group/{name}` — delete group

All user endpoints use `user:read` or `user:write`. Group endpoints
use the same permissions.

Request/response schemas for users and groups.

- [ ] **Step 2: Generate code and implement user handlers**

Category `"node"`. User operations use `job.OperationUser*`.
User list/get use `JobClient.Query`. User create/update/delete/
password use `JobClient.Modify`.

- [ ] **Step 3: Create handler.go (self-registration)**

- [ ] **Step 4: Write tests with RBAC for user handlers**

- [ ] **Step 5: Wire in controller_setup.go**

```bash
go build ./...
go test ./internal/controller/api/node/user/... ./cmd/...
git commit -m "feat(user): add user OpenAPI spec and API handlers"
```

---

## Task 7: Group API Handlers

**Files:**
- Create: `internal/controller/api/node/user/group_list_get.go`
- Create: `internal/controller/api/node/user/group_get.go`
- Create: `internal/controller/api/node/user/group_create.go`
- Create: `internal/controller/api/node/user/group_update.go`
- Create: `internal/controller/api/node/user/group_delete.go`
- Create: test files for each handler

Group handlers live in the same `api/node/user/` package as user
handlers — they share the same OpenAPI spec and handler struct.

- [ ] **Step 1: Implement group handlers with broadcast support**

Category `"node"`. Group operations use `job.OperationGroup*`.

- [ ] **Step 2: Write tests with RBAC**

- [ ] **Step 3: Verify**

```bash
go test ./internal/controller/api/node/user/...
git commit -m "feat(user): add group API handlers with tests"
```

---

## Task 8: SDK Services

**Files:**
- Create: `pkg/sdk/client/user.go`
- Create: `pkg/sdk/client/user_types.go`
- Create: `pkg/sdk/client/user_public_test.go`
- Create: `pkg/sdk/client/user_types_public_test.go`
- Create: `pkg/sdk/client/group.go`
- Create: `pkg/sdk/client/group_types.go`
- Create: `pkg/sdk/client/group_public_test.go`
- Create: `pkg/sdk/client/group_types_public_test.go`
- Modify: `pkg/sdk/client/osapi.go`

Two SDK services:

**UserService:**
- `List(ctx, hostname)`
- `Get(ctx, hostname, name)`
- `Create(ctx, hostname, opts)`
- `Update(ctx, hostname, name, opts)`
- `Delete(ctx, hostname, name)`
- `ChangePassword(ctx, hostname, name, password)`

**GroupService:**
- `List(ctx, hostname)`
- `Get(ctx, hostname, name)`
- `Create(ctx, hostname, opts)`
- `Update(ctx, hostname, name, opts)`
- `Delete(ctx, hostname, name)`

Wire both into `osapi.go`: `User *UserService`, `Group *GroupService`.

Run `just generate` before creating SDK services.

- [ ] **Step 1: Create user SDK service + types + tests**
- [ ] **Step 2: Create group SDK service + types + tests**
- [ ] **Step 3: Wire into osapi.go**
- [ ] **Step 4: Verify**

```bash
go test ./pkg/sdk/client/...
git commit -m "feat(user): add SDK services with tests"
```

---

## Task 9: CLI Commands

**Files:**
- Create: `cmd/client_node_user.go` — parent
- Create: `cmd/client_node_user_list.go`
- Create: `cmd/client_node_user_get.go`
- Create: `cmd/client_node_user_create.go`
- Create: `cmd/client_node_user_update.go`
- Create: `cmd/client_node_user_delete.go`
- Create: `cmd/client_node_user_password.go`
- Create: `cmd/client_node_group.go` — parent
- Create: `cmd/client_node_group_list.go`
- Create: `cmd/client_node_group_get.go`
- Create: `cmd/client_node_group_create.go`
- Create: `cmd/client_node_group_update.go`
- Create: `cmd/client_node_group_delete.go`

- [ ] **Step 1: Create user CLI commands**

User list: table fields NAME, UID, GID, HOME, SHELL, GROUPS, LOCKED.
User get: flags `--name` (required). Same table.
User create: flags `--name` (required), `--uid`, `--gid`, `--home`,
  `--shell`, `--groups` (string slice), `--password`, `--system`.
User update: flags `--name` (required), `--shell`, `--home`,
  `--groups` (string slice), `--lock`, `--unlock`.
User delete: flags `--name` (required).
User password: flags `--name` (required), `--password` (required).

- [ ] **Step 2: Create group CLI commands**

Group list: table fields NAME, GID, MEMBERS.
Group get: flags `--name` (required).
Group create: flags `--name` (required), `--gid`, `--system`.
Group update: flags `--name` (required), `--members` (string slice).
Group delete: flags `--name` (required).

- [ ] **Step 3: Verify build**

```bash
go build ./...
git commit -m "feat(user): add CLI commands"
```

---

## Task 10: Docs + Examples + Integration Test

**Files:**
- Create: `examples/sdk/client/user.go`
- Create: `examples/sdk/client/group.go`
- Create: `test/integration/user_test.go`
- Create: `docs/docs/sidebar/features/user-management.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/user/user.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/user/list.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/user/get.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/user/create.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/user/update.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/user/delete.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/user/password.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/group/group.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/group/list.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/group/get.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/group/create.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/group/update.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/group/delete.md`
- Create: `docs/docs/sidebar/sdk/client/management/user.md`
- Create: `docs/docs/sidebar/sdk/client/management/group.md`
- Modify: shared docs (features table, auth, config, api guidelines,
  architecture, client.md, docusaurus.config.ts)

- [ ] **Step 1: Create SDK examples (user.go, group.go)**

User example: list users, get by name.
Group example: list groups, create a group.

- [ ] **Step 2: Create integration test**

`UserSmokeSuite` with `TestUserList` (read-only) and
`TestGroupList` (read-only). Guard writes with `skipWrite`.

- [ ] **Step 3: Create feature page + CLI docs**

Feature page: `user-management.md`. CLI docs as directories with
landing pages.

- [ ] **Step 4: Create SDK doc pages**

Under `management/`: `user.md` and `group.md`.
Add both to `client.md` Management table.

- [ ] **Step 5: Update all shared docs**

Features table, auth permissions, config roles, API guidelines
(11 endpoints), architecture feature link, docusaurus dropdowns
(Features + SDK under Management group).

- [ ] **Step 6: Regenerate and verify**

```bash
just generate
go build ./...
just go::unit
just go::unit-cov  # >= 99.9%
just go::vet
```

- [ ] **Step 7: Commit**

```bash
git commit -m "feat(user): add docs, SDK examples, and integration tests"
```
