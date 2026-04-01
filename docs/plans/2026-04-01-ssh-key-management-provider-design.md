# SSH Key Management Provider Design

## Overview

Add SSH authorized key management to OSAPI. List, add, and remove
SSH public keys in a user's `~/.ssh/authorized_keys` file. Extends
the existing user provider — no new provider package or
permissions. Manages any key regardless of who added it.

## Architecture

Extends the existing user provider at
`internal/provider/node/user/`.

- **Category**: `node`
- **Path prefix**: `/node/{hostname}/user/{name}/ssh-key`
- **Permissions**: `user:read` (list), `user:write` (add, remove)
- **Provider type**: direct-write (avfs.VFS)

No state tracking, no file.Deployer. The provider reads and
writes `authorized_keys` directly. The orchestrator is
responsible for desired-state management.

## Provider Interface Additions

Added to the existing `user.Provider` interface:

```go
ListKeys(ctx context.Context, username string) ([]SSHKey, error)
AddKey(ctx context.Context, username string, key SSHKey) (*SSHKeyResult, error)
RemoveKey(ctx context.Context, username string, fingerprint string) (*SSHKeyResult, error)
```

## Data Types

```go
type SSHKey struct {
    Type        string `json:"type"`
    Fingerprint string `json:"fingerprint"`
    Comment     string `json:"comment,omitempty"`
}

type SSHKeyResult struct {
    Changed bool `json:"changed"`
}
```

## Debian Implementation

The provider resolves the user's home directory from
`/etc/passwd` (already parsed by the user provider), then
operates on `~/.ssh/authorized_keys`.

- **ListKeys**: Read `authorized_keys`, parse each non-empty,
  non-comment line into type + base64 key + comment. Compute
  SHA256 fingerprint from decoded key bytes. Return all entries.
- **AddKey**: Check if key already exists by fingerprint. If
  present, return `changed: false`. Otherwise append the raw
  public key line. Create `~/.ssh/` (mode `0700`) and
  `authorized_keys` (mode `0600`) if they don't exist. Set
  ownership to the target user via `exec.Manager`
  (`chown user:user`).
- **RemoveKey**: Read file, filter out the line matching the
  fingerprint, rewrite file. Return `changed: false` if
  fingerprint not found. Return error if file doesn't exist.

## Platform Implementations

| Platform | Implementation         |
| -------- | ---------------------- |
| Debian   | Direct file read/write |
| Darwin   | ErrUnsupported         |
| Linux    | ErrUnsupported         |

## Container Behavior

No container check — SSH key management works in containers.

## API Endpoints

| Method   | Path                                                 | Permission   | Description          |
| -------- | ---------------------------------------------------- | ------------ | -------------------- |
| `GET`    | `/node/{hostname}/user/{name}/ssh-key`               | `user:read`  | List authorized keys |
| `POST`   | `/node/{hostname}/user/{name}/ssh-key`               | `user:write` | Add a key            |
| `DELETE` | `/node/{hostname}/user/{name}/ssh-key/{fingerprint}` | `user:write` | Remove a key         |

All endpoints support broadcast targeting.

### POST Request Body

```json
{
  "key": "ssh-ed25519 AAAA... user@host"
}
```

The full public key line as it would appear in
`authorized_keys`.

### Response Shape (List)

```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "keys": [
      {
        "type": "ssh-ed25519",
        "fingerprint": "SHA256:abc123...",
        "comment": "john@laptop"
      }
    ]
  }]
}
```

### Response Shape (Add/Remove)

```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "changed": true
  }]
}
```

## SDK

```go
client.User.ListKeys(ctx, host, username)
client.User.AddKey(ctx, host, username, opts)
client.User.RemoveKey(ctx, host, username, fingerprint)
```

`SSHKeyAddOpts` struct with `Key` field (the full public key
string).

## CLI

```bash
osapi client node user ssh-key list --target web-01 --name john
osapi client node user ssh-key add --target web-01 --name john \
  --key "ssh-ed25519 AAAA... john@laptop"
osapi client node user ssh-key remove --target web-01 --name john \
  --fingerprint "SHA256:abc123..."
```

## Permissions

Reuses existing permissions — no new permissions needed.

- `user:read` — list keys
- `user:write` — add and remove keys

These are already in all built-in roles.
