# SSH Key Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add SSH authorized key management to OSAPI — list, add,
and remove SSH public keys in a user's `~/.ssh/authorized_keys`
file by extending the existing user provider.

**Architecture:** Extends `internal/provider/node/user/` with three
new methods (ListKeys, AddKey, RemoveKey). New SSH key endpoints
added to the existing user OpenAPI spec. Operations dispatched via
a new `sshKey` case in the node processor. Reuses existing
`user:read`/`user:write` permissions. No new provider package,
agent category, or permissions needed.

**Tech Stack:** Go, avfs.VFS, crypto/sha256 for fingerprints,
encoding/base64 for key decoding, oapi-codegen strict-server

---

## File Structure

### Provider Layer

- Modify: `internal/provider/node/user/types.go` — add SSHKey,
  SSHKeyResult types + 3 methods to Provider interface
- Create: `internal/provider/node/user/debian_ssh_key.go` — Debian
  implementation (list/add/remove authorized_keys)
- Modify: `internal/provider/node/user/darwin.go` — add 3 stub
  methods
- Modify: `internal/provider/node/user/linux.go` — add 3 stub
  methods
- Test: `internal/provider/node/user/debian_ssh_key_public_test.go`
- Modify: `internal/provider/node/user/darwin_public_test.go` — add
  stub tests
- Modify: `internal/provider/node/user/linux_public_test.go` — add
  stub tests

### Agent Layer

- Create: `internal/agent/processor_ssh_key.go` — SSH key operation
  dispatcher
- Modify: `internal/agent/processor.go` — add `sshKey` case to
  NewNodeProcessor
- Test: `internal/agent/processor_ssh_key_public_test.go`

### Operations

- Modify: `pkg/sdk/client/operations.go` — add SSH key operation
  constants
- Modify: `internal/job/types.go` — add SSH key operation aliases

### API Layer

- Modify: `internal/controller/api/node/user/gen/api.yaml` — add
  3 ssh-key endpoints + schemas
- Create:
  `internal/controller/api/node/user/ssh_key_list_get.go` — list
  handler
- Create:
  `internal/controller/api/node/user/ssh_key_add_post.go` — add
  handler
- Create:
  `internal/controller/api/node/user/ssh_key_remove_delete.go` —
  remove handler
- Test:
  `internal/controller/api/node/user/ssh_key_list_get_public_test.go`
- Test:
  `internal/controller/api/node/user/ssh_key_add_post_public_test.go`
- Test:
  `internal/controller/api/node/user/ssh_key_remove_delete_public_test.go`

### SDK Layer

- Modify: `pkg/sdk/client/user.go` — add ListKeys, AddKey,
  RemoveKey methods
- Modify: `pkg/sdk/client/user_types.go` — add SSHKey result
  types + conversions
- Modify: `pkg/sdk/client/user_public_test.go` — add tests
- Modify: `pkg/sdk/client/user_types_public_test.go` — add
  conversion tests

### CLI Layer

- Create: `cmd/client_node_user_ssh_key.go` — parent command
- Create: `cmd/client_node_user_ssh_key_list.go` — list
  subcommand
- Create: `cmd/client_node_user_ssh_key_add.go` — add subcommand
- Create: `cmd/client_node_user_ssh_key_remove.go` — remove
  subcommand

### Documentation

- Modify: `docs/docs/sidebar/features/user-management.md` — add
  SSH key section
- Create:
  `docs/docs/sidebar/usage/cli/client/node/user/ssh-key.md` — CLI
  landing
- Create:
  `docs/docs/sidebar/usage/cli/client/node/user/ssh-key-list.md`
- Create:
  `docs/docs/sidebar/usage/cli/client/node/user/ssh-key-add.md`
- Create:
  `docs/docs/sidebar/usage/cli/client/node/user/ssh-key-remove.md`
- Modify: `docs/docs/sidebar/sdk/client/management/user.md` — add
  SSH key methods
- Modify: `examples/sdk/client/user.go` — add SSH key demo
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md` — add
  endpoints

---

### Task 1: Provider Types and Stubs

**Files:**
- Modify: `internal/provider/node/user/types.go`
- Modify: `internal/provider/node/user/darwin.go`
- Modify: `internal/provider/node/user/linux.go`
- Modify: `internal/provider/node/user/darwin_public_test.go`
- Modify: `internal/provider/node/user/linux_public_test.go`

- [ ] **Step 1: Add types to types.go**

Add to `internal/provider/node/user/types.go`:

```go
// SSHKey represents an SSH authorized key entry.
type SSHKey struct {
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment,omitempty"`
}

// SSHKeyResult represents the result of an SSH key mutation.
type SSHKeyResult struct {
	Changed bool `json:"changed"`
}
```

Add 3 methods to the Provider interface:

```go
	// ListKeys returns SSH authorized keys for a user.
	ListKeys(ctx context.Context, username string) ([]SSHKey, error)
	// AddKey adds an SSH authorized key for a user.
	AddKey(ctx context.Context, username string, key SSHKey) (*SSHKeyResult, error)
	// RemoveKey removes an SSH authorized key by fingerprint.
	RemoveKey(ctx context.Context, username string, fingerprint string) (*SSHKeyResult, error)
```

- [ ] **Step 2: Add stub methods to darwin.go and linux.go**

Add to both `darwin.go` and `linux.go`:

```go
// ListKeys returns ErrUnsupported on Darwin/Linux.
func (d *Darwin) ListKeys(
	_ context.Context,
	_ string,
) ([]SSHKey, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// AddKey returns ErrUnsupported on Darwin/Linux.
func (d *Darwin) AddKey(
	_ context.Context,
	_ string,
	_ SSHKey,
) (*SSHKeyResult, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}

// RemoveKey returns ErrUnsupported on Darwin/Linux.
func (d *Darwin) RemoveKey(
	_ context.Context,
	_ string,
	_ string,
) (*SSHKeyResult, error) {
	return nil, fmt.Errorf("user: %w", provider.ErrUnsupported)
}
```

(Same for Linux struct.)

- [ ] **Step 3: Add stub tests**

Add test cases to the existing test tables in
`darwin_public_test.go` and `linux_public_test.go` for
ListKeys, AddKey, and RemoveKey all returning
ErrUnsupported.

- [ ] **Step 4: Regenerate mocks**

Run: `go generate ./internal/provider/node/user/mocks/...`

- [ ] **Step 5: Run tests**

Run: `go test -v ./internal/provider/node/user/...`
Expected: all pass, new stub tests included

- [ ] **Step 6: Commit**

```bash
git add internal/provider/node/user/
git commit -m "feat(user): add SSH key types and platform stubs"
```

---

### Task 2: Debian SSH Key Implementation

**Files:**
- Create: `internal/provider/node/user/debian_ssh_key.go`
- Test: `internal/provider/node/user/debian_ssh_key_public_test.go`

- [ ] **Step 1: Write tests**

Create `debian_ssh_key_public_test.go` with testify/suite.
Use `memfs.New()` for filesystem and gomock for exec.Manager.

Set up a memfs with `/etc/passwd` containing:
```
root:x:0:0:root:/root:/bin/bash
john:x:1000:1000:John:/home/john:/bin/bash
```

**TestListKeys** — table-driven:
- success (authorized_keys with 2 keys, verify type + fingerprint
  + comment)
- user not found in /etc/passwd → error
- no authorized_keys file → empty list, no error
- empty authorized_keys → empty list
- lines with comments and blank lines skipped
- malformed key line skipped (logged as debug)

**TestAddKey** — table-driven:
- success (creates .ssh dir + file, appends key)
- key already exists (same fingerprint) → changed: false
- user not found → error
- creates .ssh dir with 0700 if missing
- creates authorized_keys with 0600 if missing
- appends to existing file

**TestRemoveKey** — table-driven:
- success (rewrites file without matching key)
- fingerprint not found → changed: false
- user not found → error
- no authorized_keys file → changed: false
- file becomes empty after removal (still valid)

- [ ] **Step 2: Implement debian_ssh_key.go**

```go
package user

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
)

// ListKeys returns SSH authorized keys for a user.
func (d *Debian) ListKeys(
	_ context.Context,
	username string,
) ([]SSHKey, error) {
	d.logger.Debug("executing user.ListKeys",
		slog.String("username", username),
	)

	home, err := d.userHomeDir(username)
	if err != nil {
		return nil, fmt.Errorf("ssh key: list: %w", err)
	}

	authKeysPath := home + "/.ssh/authorized_keys"

	content, err := d.fs.ReadFile(authKeysPath)
	if err != nil {
		// No file = no keys, not an error.
		return []SSHKey{}, nil
	}

	return parseAuthorizedKeys(string(content), d.logger), nil
}

// AddKey adds an SSH authorized key for a user.
func (d *Debian) AddKey(
	_ context.Context,
	username string,
	key SSHKey,
) (*SSHKeyResult, error) {
	d.logger.Debug("executing user.AddKey",
		slog.String("username", username),
	)

	home, err := d.userHomeDir(username)
	if err != nil {
		return nil, fmt.Errorf("ssh key: add: %w", err)
	}

	sshDir := home + "/.ssh"
	authKeysPath := sshDir + "/authorized_keys"

	// Ensure .ssh directory exists.
	if err := d.fs.MkdirAll(sshDir, 0o700); err != nil {
		return nil, fmt.Errorf(
			"ssh key: create .ssh dir: %w", err)
	}

	// Read existing keys to check for duplicates.
	existing, _ := d.fs.ReadFile(authKeysPath)
	existingKeys := parseAuthorizedKeys(
		string(existing), d.logger)

	for _, ek := range existingKeys {
		if ek.Fingerprint == key.Fingerprint {
			return &SSHKeyResult{Changed: false}, nil
		}
	}

	// Build the key line from the SSHKey fields.
	keyLine := key.Type + " " +
		base64.StdEncoding.EncodeToString(/* raw key bytes */)
	// Actually, the API receives the full key line in a
	// dedicated field. See the AddKey handler — it passes
	// the raw key line. The provider should store the raw
	// public key line.

	// Append key to file.
	f, err := d.fs.OpenFile(
		authKeysPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o600,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"ssh key: open authorized_keys: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(
		[]byte(key.RawLine + "\n"),
	); err != nil {
		return nil, fmt.Errorf(
			"ssh key: write key: %w", err)
	}

	// Set ownership.
	if _, err := d.execManager.RunCmd(
		"chown",
		[]string{"-R", username + ":" + username, sshDir},
	); err != nil {
		d.logger.Warn("failed to set .ssh ownership",
			slog.String("error", err.Error()),
		)
	}

	return &SSHKeyResult{Changed: true}, nil
}

// RemoveKey removes an SSH authorized key by fingerprint.
func (d *Debian) RemoveKey(
	_ context.Context,
	username string,
	fingerprint string,
) (*SSHKeyResult, error) {
	d.logger.Debug("executing user.RemoveKey",
		slog.String("username", username),
		slog.String("fingerprint", fingerprint),
	)

	home, err := d.userHomeDir(username)
	if err != nil {
		return nil, fmt.Errorf("ssh key: remove: %w", err)
	}

	authKeysPath := home + "/.ssh/authorized_keys"

	content, err := d.fs.ReadFile(authKeysPath)
	if err != nil {
		return &SSHKeyResult{Changed: false}, nil
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	found := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}

		fp := fingerprintFromLine(trimmed)
		if fp == fingerprint {
			found = true
			continue // skip this line
		}
		newLines = append(newLines, line)
	}

	if !found {
		return &SSHKeyResult{Changed: false}, nil
	}

	newContent := strings.Join(newLines, "\n")
	if err := d.fs.WriteFile(
		authKeysPath, []byte(newContent), 0o600,
	); err != nil {
		return nil, fmt.Errorf(
			"ssh key: write authorized_keys: %w", err)
	}

	return &SSHKeyResult{Changed: true}, nil
}

// userHomeDir resolves a user's home directory from
// /etc/passwd.
func (d *Debian) userHomeDir(
	username string,
) (string, error) {
	content, err := d.fs.ReadFile("/etc/passwd")
	if err != nil {
		return "", fmt.Errorf(
			"read /etc/passwd: %w", err)
	}

	for _, line := range strings.Split(
		string(content), "\n") {
		fields := strings.Split(line, ":")
		if len(fields) >= 6 && fields[0] == username {
			return fields[5], nil
		}
	}

	return "", fmt.Errorf("user %q not found", username)
}

// parseAuthorizedKeys parses an authorized_keys file content
// into SSHKey entries.
func parseAuthorizedKeys(
	content string,
	logger *slog.Logger,
) []SSHKey {
	var keys []SSHKey

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			logger.Debug("skipping malformed key line",
				slog.String("line", line),
			)
			continue
		}

		keyType := parts[0]
		keyData := parts[1]
		comment := ""
		if len(parts) >= 3 {
			comment = strings.Join(parts[2:], " ")
		}

		fp := computeFingerprint(keyData)
		if fp == "" {
			logger.Debug("skipping key with invalid base64",
				slog.String("line", line),
			)
			continue
		}

		keys = append(keys, SSHKey{
			Type:        keyType,
			Fingerprint: fp,
			Comment:     comment,
		})
	}

	return keys
}

// computeFingerprint computes SHA256 fingerprint from base64-
// encoded key data.
func computeFingerprint(
	keyData string,
) string {
	decoded, err := base64.StdEncoding.DecodeString(keyData)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(decoded)

	return "SHA256:" +
		base64.RawStdEncoding.EncodeToString(hash[:])
}

// fingerprintFromLine extracts fingerprint from a key line.
func fingerprintFromLine(
	line string,
) string {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return ""
	}

	return computeFingerprint(parts[1])
}
```

**IMPORTANT**: The SSHKey type needs a `RawLine` field to store
the full public key string for AddKey. Update the types:

```go
type SSHKey struct {
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment,omitempty"`
	RawLine     string `json:"raw_line,omitempty"`
}
```

The API handler populates `RawLine` from the POST body's `key`
field. The provider uses `RawLine` to append to
`authorized_keys`. ListKeys does NOT populate `RawLine` (we
don't expose raw key data in list responses — just type,
fingerprint, comment).

- [ ] **Step 3: Run tests**

Run: `go test -v ./internal/provider/node/user/...`
Expected: all pass

- [ ] **Step 4: Verify 100% coverage on new file**

```bash
go test -coverprofile=/tmp/ssh.cov \
  ./internal/provider/node/user/... && \
  go tool cover -func=/tmp/ssh.cov | \
  grep "debian_ssh_key"
```

All functions must be 100%.

- [ ] **Step 5: Commit**

```bash
git add internal/provider/node/user/
git commit -m "feat(user): add SSH key management to debian provider"
```

---

### Task 3: Operations and Agent Processor

**Files:**
- Modify: `pkg/sdk/client/operations.go`
- Modify: `internal/job/types.go`
- Create: `internal/agent/processor_ssh_key.go`
- Modify: `internal/agent/processor.go` — add `sshKey` case
- Test: `internal/agent/processor_ssh_key_public_test.go`

- [ ] **Step 1: Add operation constants**

In `pkg/sdk/client/operations.go`, add after User operations:

```go
// SSH Key operations.
const (
	OpSSHKeyList   JobOperation = "node.sshKey.list"
	OpSSHKeyAdd    JobOperation = "node.sshKey.add"
	OpSSHKeyRemove JobOperation = "node.sshKey.remove"
)
```

In `internal/job/types.go`, add corresponding aliases:

```go
// SSH Key operations.
const (
	OperationSSHKeyList   = client.OpSSHKeyList
	OperationSSHKeyAdd    = client.OpSSHKeyAdd
	OperationSSHKeyRemove = client.OpSSHKeyRemove
)
```

- [ ] **Step 2: Write processor tests**

Create `internal/agent/processor_ssh_key_public_test.go`.
The processor dispatches to the existing `userProvider` (same
as user/group operations). Test via `NewNodeProcessor`.

**TestProcessSSHKeyOperation** — dispatch-level table:
- nil user provider → error
- invalid operation format
- unsupported sub-operation

**TestProcessSSHKeyList** — table-driven:
- success (returns keys)
- unmarshal error (invalid JSON)
- provider error

**TestProcessSSHKeyAdd** — table-driven:
- success
- unmarshal error
- provider error

**TestProcessSSHKeyRemove** — table-driven:
- success
- unmarshal error
- provider error

One suite method per function, ALL scenarios as table rows.

- [ ] **Step 3: Implement processor_ssh_key.go**

```go
func processSshKeyOperation(
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if userProvider == nil {
		return nil, fmt.Errorf(
			"user provider not available")
	}

	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf(
			"invalid sshKey operation: %s",
			jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processSshKeyList(
			ctx, userProvider, logger, jobRequest)
	case "add":
		return processSshKeyAdd(
			ctx, userProvider, logger, jobRequest)
	case "remove":
		return processSshKeyRemove(
			ctx, userProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf(
			"unsupported sshKey operation: %s",
			jobRequest.Operation)
	}
}
```

Each sub-handler unmarshals username (and key data for add,
fingerprint for remove) from `jobRequest.Data`, calls the
provider, and marshals the result.

- [ ] **Step 4: Wire into node processor**

In `internal/agent/processor.go`, add case to the
`NewNodeProcessor` switch:

```go
		case "sshKey":
			return processSshKeyOperation(
				userProvider, logger, req)
```

- [ ] **Step 5: Run tests and verify coverage**

```bash
go test -v ./internal/agent/...
go build ./...
go test -coverprofile=/tmp/ssh_proc.cov \
  ./internal/agent/... && \
  go tool cover -func=/tmp/ssh_proc.cov | \
  grep "processor_ssh_key"
```

- [ ] **Step 6: Commit**

```bash
git add pkg/sdk/client/operations.go \
  internal/job/types.go \
  internal/agent/processor_ssh_key.go \
  internal/agent/processor_ssh_key_public_test.go \
  internal/agent/processor.go
git commit -m "feat(user): add SSH key operations and agent processor"
```

---

### Task 4: OpenAPI Spec Update and Code Generation

**Files:**
- Modify: `internal/controller/api/node/user/gen/api.yaml`

- [ ] **Step 1: Add endpoints to existing user OpenAPI spec**

Add to `internal/controller/api/node/user/gen/api.yaml` after
the password endpoint section:

```yaml
  # -- SSH Key management ------------------------------------------------

  /node/{hostname}/user/{name}/ssh-key:
    get:
      summary: List SSH authorized keys
      description: >
        List SSH authorized keys for a user on the target node.
      tags:
        - user_operations
      operationId: GetNodeUserSshKey
      security:
        - BearerAuth:
            - user:read
      parameters:
        - $ref: '#/components/parameters/Hostname'
        - $ref: '#/components/parameters/UserName'
      responses:
        '200':
          description: List of SSH authorized keys.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SSHKeyCollectionResponse'
        '401': ...
        '403': ...
        '500': ...

    post:
      summary: Add SSH authorized key
      description: >
        Add an SSH authorized key for a user on the target node.
      tags:
        - user_operations
      operationId: PostNodeUserSshKey
      security:
        - BearerAuth:
            - user:write
      parameters:
        - $ref: '#/components/parameters/Hostname'
        - $ref: '#/components/parameters/UserName'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/SSHKeyAddRequest'
      responses:
        '200':
          description: Key added.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SSHKeyMutationResponse'
        '400': ...
        '401': ...
        '403': ...
        '500': ...

  /node/{hostname}/user/{name}/ssh-key/{fingerprint}:
    delete:
      summary: Remove SSH authorized key
      description: >
        Remove an SSH authorized key by fingerprint.
      tags:
        - user_operations
      operationId: DeleteNodeUserSshKey
      security:
        - BearerAuth:
            - user:write
      parameters:
        - $ref: '#/components/parameters/Hostname'
        - $ref: '#/components/parameters/UserName'
        - $ref: '#/components/parameters/SSHKeyFingerprint'
      responses:
        '200':
          description: Key removed.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SSHKeyMutationResponse'
        '401': ...
        '403': ...
        '500': ...
```

Add schemas:

```yaml
    SSHKeyAddRequest:
      type: object
      required:
        - key
      properties:
        key:
          type: string
          description: >
            Full SSH public key line (e.g.,
            "ssh-ed25519 AAAA... user@host").
          x-oapi-codegen-extra-tags:
            validate: required,min=1

    SSHKeyInfo:
      type: object
      properties:
        type:
          type: string
          example: "ssh-ed25519"
        fingerprint:
          type: string
          example: "SHA256:abc123..."
        comment:
          type: string
          example: "john@laptop"

    SSHKeyEntry:
      type: object
      properties:
        hostname:
          type: string
        status:
          type: string
          enum: [ok, failed, skipped]
        keys:
          type: array
          items:
            $ref: '#/components/schemas/SSHKeyInfo'
        error:
          type: string
      required:
        - hostname
        - status

    SSHKeyMutationEntry:
      type: object
      properties:
        hostname:
          type: string
        status:
          type: string
          enum: [ok, failed, skipped]
        changed:
          type: boolean
        error:
          type: string
      required:
        - hostname
        - status

    SSHKeyCollectionResponse:
      type: object
      properties:
        job_id:
          type: string
          format: uuid
        results:
          type: array
          items:
            $ref: '#/components/schemas/SSHKeyEntry'
      required:
        - results

    SSHKeyMutationResponse:
      type: object
      properties:
        job_id:
          type: string
          format: uuid
        results:
          type: array
          items:
            $ref: '#/components/schemas/SSHKeyMutationEntry'
      required:
        - results
```

Add parameter:

```yaml
    SSHKeyFingerprint:
      name: fingerprint
      in: path
      required: true
      description: SSH key SHA256 fingerprint.
      x-oapi-codegen-extra-tags:
        validate: required,min=1
      schema:
        type: string
        minLength: 1
```

- [ ] **Step 2: Generate code and rebuild**

```bash
go generate ./internal/controller/api/node/user/gen/...
just generate
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/controller/api/node/user/gen/ \
  internal/controller/api/gen/ \
  pkg/sdk/client/gen/
git commit -m "feat(user): add SSH key endpoints to OpenAPI spec"
```

---

### Task 5: API Handler Implementation

**Files:**
- Create: `internal/controller/api/node/user/ssh_key_list_get.go`
- Create: `internal/controller/api/node/user/ssh_key_add_post.go`
- Create: `internal/controller/api/node/user/ssh_key_remove_delete.go`
- Test: all 3 `*_public_test.go` files

- [ ] **Step 1: Implement list handler**

`GetNodeUserSshKey` method on the existing `User` handler
struct:
- Validate hostname
- username from `request.Name`
- Query with category `"node"`, operation
  `job.OperationSSHKeyList`, data `{"username": username}`
- Parse response: unmarshal `[]userProv.SSHKey`, convert to
  `[]gen.SSHKeyInfo`
- Broadcast support

- [ ] **Step 2: Implement add handler**

`PostNodeUserSshKey`:
- Validate hostname, body (`key` field)
- Parse the raw key line to extract type, fingerprint, comment
- Build `userProv.SSHKey{Type, Fingerprint, Comment, RawLine}`
- Modify with `job.OperationSSHKeyAdd`, data includes
  `username` + the SSHKey struct
- Parse mutation response

- [ ] **Step 3: Implement remove handler**

`DeleteNodeUserSshKey`:
- Validate hostname
- fingerprint from `request.Fingerprint`
- Modify with `job.OperationSSHKeyRemove`, data
  `{"username": username, "fingerprint": fingerprint}`
- Parse mutation response

- [ ] **Step 4: Write tests**

Each handler test file needs: success, skipped, broadcast,
validation error, job error, HTTP wiring, RBAC (401/403/200).
One suite method per handler, all scenarios as table rows.

- [ ] **Step 5: Run tests and verify coverage**

```bash
go test -v ./internal/controller/api/node/user/...
go test -coverprofile=/tmp/ssh_h.cov \
  ./internal/controller/api/node/user/... && \
  go tool cover -func=/tmp/ssh_h.cov | \
  grep "ssh_key" | grep -v "100.0%"
```

- [ ] **Step 6: Commit**

```bash
git add internal/controller/api/node/user/
git commit -m "feat(user): add SSH key API handlers with broadcast support"
```

---

### Task 6: SDK Service Extension

**Files:**
- Modify: `pkg/sdk/client/user.go`
- Modify: `pkg/sdk/client/user_types.go`
- Modify: `pkg/sdk/client/user_public_test.go`
- Modify: `pkg/sdk/client/user_types_public_test.go`

- [ ] **Step 1: Add types**

In `user_types.go`, add:

```go
type SSHKeyInfoResult struct {
	Hostname string        `json:"hostname"`
	Status   string        `json:"status"`
	Keys     []SSHKeyInfo  `json:"keys,omitempty"`
	Error    string        `json:"error,omitempty"`
}

type SSHKeyInfo struct {
	Type        string `json:"type,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

type SSHKeyMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

type SSHKeyAddOpts struct {
	Key string
}
```

Add conversion functions.

- [ ] **Step 2: Add methods to UserService**

In `user.go`, add:

```go
func (s *UserService) ListKeys(
	ctx context.Context,
	hostname string,
	username string,
) (*Response[Collection[SSHKeyInfoResult]], error)

func (s *UserService) AddKey(
	ctx context.Context,
	hostname string,
	username string,
	opts SSHKeyAddOpts,
) (*Response[Collection[SSHKeyMutationResult]], error)

func (s *UserService) RemoveKey(
	ctx context.Context,
	hostname string,
	username string,
	fingerprint string,
) (*Response[Collection[SSHKeyMutationResult]], error)
```

- [ ] **Step 3: Regenerate SDK client**

```bash
go generate ./pkg/sdk/client/gen/...
```

- [ ] **Step 4: Write tests**

Add tests to existing test files (or create new
`user_ssh_key_public_test.go` / `user_ssh_key_types_public_test.go`
if the existing files are already large). Follow existing
patterns with httptest.Server.

- [ ] **Step 5: Verify 100% coverage**

```bash
go test -coverprofile=/tmp/ssh_sdk.cov \
  ./pkg/sdk/client/... && \
  go tool cover -func=/tmp/ssh_sdk.cov | \
  grep "user" | grep -v "100.0%"
```

- [ ] **Step 6: Commit**

```bash
git add pkg/sdk/client/
git commit -m "feat(user): add SSH key SDK methods with tests"
```

---

### Task 7: CLI Commands

**Files:**
- Create: `cmd/client_node_user_ssh_key.go`
- Create: `cmd/client_node_user_ssh_key_list.go`
- Create: `cmd/client_node_user_ssh_key_add.go`
- Create: `cmd/client_node_user_ssh_key_remove.go`

- [ ] **Step 1: Create parent command**

```go
var clientNodeUserSshKeyCmd = &cobra.Command{
	Use:   "ssh-key",
	Short: "Manage SSH authorized keys",
}

func init() {
	clientNodeUserCmd.AddCommand(clientNodeUserSshKeyCmd)
}
```

Wait — check whether `clientNodeUserCmd` exists. Look at
`cmd/client_node_user.go` for the parent.

- [ ] **Step 2: Create list subcommand**

Flags: `--name` (username, required)
- Calls `sdkClient.User.ListKeys(ctx, host, name)`
- Table headers: `TYPE`, `FINGERPRINT`, `COMMENT`
- Uses `BuildBroadcastTable`

- [ ] **Step 3: Create add subcommand**

Flags: `--name` (required), `--key` (required, full public
key line)
- Calls `sdkClient.User.AddKey(ctx, host, name, opts)`
- Uses `BuildMutationTable` with headers `CHANGED`

- [ ] **Step 4: Create remove subcommand**

Flags: `--name` (required), `--fingerprint` (required)
- Calls `sdkClient.User.RemoveKey(ctx, host, name, fp)`
- Uses `BuildMutationTable`

- [ ] **Step 5: Verify build**

```bash
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add cmd/client_node_user_ssh_key*.go
git commit -m "feat(user): add SSH key CLI commands"
```

---

### Task 8: Documentation

**Files:**
- Modify: `docs/docs/sidebar/features/user-management.md`
- Create: CLI doc pages for ssh-key commands
- Modify: `docs/docs/sidebar/sdk/client/management/user.md`
- Modify: `examples/sdk/client/user.go`
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md`

- [ ] **Step 1: Update feature page**

Add SSH Key Management section to
`docs/docs/sidebar/features/user-management.md`:
- How It Works (list, add, remove)
- Add to Operations table
- Add CLI examples for ssh-key subcommands
- Note: uses existing `user:read`/`user:write` permissions

- [ ] **Step 2: Create CLI doc pages**

Create landing page + list.md, add.md, remove.md under
`docs/docs/sidebar/usage/cli/client/node/user/`.

- [ ] **Step 3: Update SDK doc**

Add ListKeys, AddKey, RemoveKey to the user SDK doc page
with code examples and result type tables.

- [ ] **Step 4: Update SDK example**

Add SSH key demo to `examples/sdk/client/user.go`.

- [ ] **Step 5: Update api-guidelines**

Add endpoint rows:
```
| `/node/{hostname}/user/{name}/ssh-key`                | User |
| `/node/{hostname}/user/{name}/ssh-key/{fingerprint}`  | User |
```

- [ ] **Step 6: Commit**

```bash
git add docs/ examples/
git commit -m "docs: add SSH key management to user docs and SDK example"
```

---

### Task 9: Integration Test and Final Verification

**Files:**
- Modify or create: `test/integration/user_test.go` (add SSH
  key tests)

- [ ] **Step 1: Add integration test**

Add SSH key list test to the existing user integration test
file (or create new if it doesn't exist). Test:
- `osapi client node user ssh-key list --target _any --name root --json`

- [ ] **Step 2: Run full suite**

```bash
just generate
go build ./...
just go::unit
just go::vet
```

- [ ] **Step 3: Verify coverage**

```bash
go test -coverprofile=/tmp/ssh_all.cov \
  ./internal/provider/node/user/... \
  ./internal/agent/... \
  ./internal/controller/api/node/user/... \
  ./pkg/sdk/client/...
go tool cover -func=/tmp/ssh_all.cov | \
  grep "ssh_key\|ssh_key" | grep -v "100.0%" | \
  grep -v "mocks\|gen/"
```

- [ ] **Step 4: Commit any fixes**

```bash
git add -A
git commit -m "chore(user): fix formatting and lint"
```
