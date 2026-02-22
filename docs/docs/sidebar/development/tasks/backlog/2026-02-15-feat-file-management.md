---
title: File and directory management
status: backlog
created: 2026-02-15
updated: 2026-02-19
---

## Objective

Add file and directory management in two parts: a content-addressed **blob
store** for getting files onto the system, and **file operations** for managing
files on disk. The blob store is a shared primitive — any future feature that
needs to push files to workers (configs, certs, packages) uses the same
upload-once, reference-by-SHA mechanism.

## Part 1: Blob Store (`/blobs`)

Content-addressed file storage with pluggable backends. Files are immutable —
same SHA = same content, no re-upload needed.

### API Endpoints

```
POST   /blobs                 - Upload file to managed store (default: NATS)
POST   /blobs/register        - Register external reference (s3://, file://, https://)
HEAD   /blobs/{sha256}        - Check if blob exists (no download)
GET    /blobs/{sha256}        - Download blob content (managed only)
DELETE /blobs/{sha256}        - Remove blob (admin only)
GET    /blobs                 - List blobs (with metadata, pagination)
```

### Two Modes

**Managed (default):** `POST /blobs` uploads bytes into the configured backend
(NATS Object Store by default). The API server receives the file and stores it.
Good for config files, scripts, small binaries.

**External reference:** `POST /blobs/register` records a pointer to a file that
already lives somewhere workers can reach — S3 bucket, NFS mount, HTTPS URL. No
bytes are transferred through OSAPI. Good for ISOs, large packages, anything
already hosted. Workers resolve the source at execution time.

```go
type BlobRef struct {
    SHA256 string // content hash (required for both modes)
    Source string // "managed" | "s3://..." | "file:///..." | "https://..."
    Size   int64
}
```

### Storage Backend Interface

```go
type BlobStore interface {
    Put(ctx context.Context, reader io.Reader, metadata BlobMetadata) (sha256 string, err error)
    Get(ctx context.Context, sha256 string) (io.ReadCloser, BlobMetadata, error)
    Exists(ctx context.Context, sha256 string) (bool, error)
    Delete(ctx context.Context, sha256 string) error
    List(ctx context.Context, opts ListOptions) ([]BlobMetadata, error)
}

type BlobMetadata struct {
    SHA256    string
    Size      int64
    Filename  string    // original filename (informational)
    MIMEType  string    // detected or provided
    CreatedAt time.Time
    Backend   string    // "nats", "s3", "fs", etc.
}
```

### Backends

| Backend               | Config key            | Best for                                                   | Notes                                              |
| --------------------- | --------------------- | ---------------------------------------------------------- | -------------------------------------------------- |
| **NATS Object Store** | `blobs.backend: nats` | Small-medium files (<100MB), single-node or small clusters | Built-in, no extra infra. Uses JetStream chunking. |
| **S3-compatible**     | `blobs.backend: s3`   | Large files, existing cloud infra                          | AWS S3, MinIO, R2, etc.                            |
| **Filesystem**        | `blobs.backend: fs`   | NAS/NFS mounts, air-gapped environments                    | Simple directory on a shared mount.                |

Backend is selected via config. Workers use the same interface to pull blobs
regardless of backend:

```yaml
blobs:
  backend: nats # or "s3" or "fs"

  nats:
    bucket: osapi-blobs
    chunk_size: 262144 # 256KB chunks

  s3:
    endpoint: s3.amazonaws.com
    bucket: osapi-blobs
    region: us-east-1
    # credentials via env: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY

  fs:
    path: /var/lib/osapi/blobs
```

### How Jobs Reference Blobs

Operations that need a file include the SHA in the job payload instead of inline
content. Workers pull the blob from the configured store when executing:

```json
{
  "type": "file.write.execute",
  "data": {
    "path": "/etc/nginx/nginx.conf",
    "source_sha256": "a1b2c3...",
    "mode": "0644",
    "owner": "root"
  }
}
```

Upload once, deploy to many workers (broadcast `_all` or label targeting). CLI
workflow:

```bash
# Upload a file to managed store (default: NATS Object Store)
osapi client blob upload nginx.conf
# → sha256: a1b2c3..., source: managed, backend: nats

# Register an external file (no upload, just a pointer)
osapi client blob register \
  --sha256 def456... \
  --source s3://my-bucket/images/ubuntu-24.04.iso \
  --size 4294967296

# Check if a blob exists (works for both managed and external)
osapi client blob check a1b2c3...
# → exists: true, size: 4096, source: managed

# Use it in a job — worker resolves source automatically
osapi client file write /etc/nginx/nginx.conf \
  --source-sha a1b2c3... --mode 0644
```

### Package

- `internal/blob/` — `BlobStore` interface + backend implementations
- `internal/api/blob/` — API handlers
- `internal/client/blob_*.go` — client wrappers
- `cmd/client_blob*.go` — CLI commands

## Part 2: File Operations (`/file`)

File and directory management on the target system. Write operations reference
blobs by SHA for file content.

### API Endpoints

```
GET    /file/stat             - Get file/directory metadata
POST   /file/read             - Read file contents (with line range)
PUT    /file/write            - Write file from blob SHA or inline content
PATCH  /file/line             - Insert/replace/remove line in file
PUT    /file/permissions      - Set owner, group, mode
POST   /file/directory        - Create directory (mkdir -p)
DELETE /file/{path}           - Delete file or directory
POST   /file/copy             - Copy file within the system
```

### Operations

- `file.stat.get` (query) — Ansible `stat` equivalent
- `file.read.get` (query)
- `file.write.execute` (modify) — write from blob SHA or small inline content
- `file.line.update` (modify) — Ansible `lineinfile` equivalent
- `file.permissions.update` (modify) — Ansible `file` with mode/owner
- `file.directory.create` (modify)
- `file.delete.execute` (modify)
- `file.copy.execute` (modify)

### Provider

- `internal/provider/system/file/`
- `stat`: path, size, mode, owner, group, modified, is_dir, is_link, checksum
  (sha256)
- `read`: line offset/limit, binary detection
- `write`: accept blob SHA or inline content + mode + owner, create parent dirs,
  optional backup
- `line`: regexp match, insertafter, insertbefore, state (present/absent) —
  mirrors Ansible lineinfile
- `permissions`: `chown`, `chmod`

## Prerequisites: nats-client Object Store Support

The `nats-client` sibling repo wraps JetStream KV, Streams, and Consumers but
does **not** yet wrap Object Store. The NATS backend for the blob store depends
on it.

**nats-server: No changes needed.** Object Store is a JetStream primitive —
automatically available when JetStream is enabled.

**nats-client: Must add Object Store wrapper.** The upstream NATS Go library
(v1.48.0) already has the full Object Store API, and mocks are already generated
from the JetStream interfaces. Just needs a wrapper layer following the `kv.go`
pattern.

### Files to add in nats-client

`pkg/client/objectstore.go` — wrapper methods:

- `CreateObjectStore(ctx, config)` — create an Object Store bucket
- `GetObjectStore(ctx, name)` — get existing bucket
- `DeleteObjectStore(ctx, name)` — delete bucket
- `PutObject(ctx, store, name, reader)` — store object (chunked)
- `GetObject(ctx, store, name)` — retrieve object
- `DeleteObject(ctx, store, name)` — delete object
- `ListObjects(ctx, store)` — list objects in a bucket
- `ObjectInfo(ctx, store, name)` — get object metadata

`pkg/client/objectstore_public_test.go` — full test coverage following the
`kv_public_test.go` table-driven pattern. Must maintain 100% coverage.

## Implementation Order

1. **nats-client Object Store wrapper** — add to sibling repo first
2. **Blob store interface + NATS backend** — `internal/blob/` using the new
   nats-client Object Store methods
3. **Blob API + CLI** — upload/check/download/list/register
4. **File operations provider** — stat, read, write (with blob integration),
   permissions
5. **File API + CLI** — endpoints and commands
6. **S3 backend** — add when needed
7. **Filesystem backend** — add when needed

## Notes

- Blob uploads go direct to the API server (not through the job system) —
  they're infrastructure, not operations
- File operations go through the job system as usual (worker executes on target
  host)
- Workers need access to the blob store to pull files — same config, same
  interface
- Path validation to prevent directory traversal attacks
- Size limits on inline content (small config snippets OK, large files must use
  blob SHA)
- Scopes: `blob:read`, `blob:write` for the store; `file:read`, `file:write` for
  operations
- `lineinfile` regex-based editing is one of Ansible's most-used features —
  worth getting right
- Consider a diff/preview endpoint before applying changes
