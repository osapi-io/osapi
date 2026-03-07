# Multipart File Upload with Streaming

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Migrate the file upload endpoint from JSON with base64-encoded content
to `multipart/form-data` with streaming to NATS Object Store. Add a
`content_type` metadata field to track file purpose (raw vs template). Increase
Object Store bucket size for large file support.

**Architecture:** Two-pass from temp file — Go's `ParseMultipartForm(32 MiB)`
spools large files to disk. First pass computes SHA-256 for idempotency check;
second pass streams to NATS via `Put(io.Reader)`. Memory is bounded at ~32 MiB
regardless of file size. Content type stored as NATS object header on upload;
deploy reads it from stored metadata.

**Tech Stack:** Go 1.25, NATS JetStream Object Store, oapi-codegen
`multipart/form-data`, testify/suite, gomock.

**Design doc:** N/A — plan originated from conversation.

---

## Step 1: Add `Put` to ObjectStoreManager Interface

**File:** `internal/api/file/types.go`

Add streaming `Put` method alongside existing `PutBytes`:

```go
Put(
    ctx context.Context,
    meta *jetstream.ObjectMeta,
    reader io.Reader,
    opts ...jetstream.ObjectOpt,
) (*jetstream.ObjectInfo, error)
```

Add `io` to imports. Keep `PutBytes` — it's still used elsewhere.

Regenerate mock:

```bash
go generate ./internal/api/file/mocks/...
```

---

## Step 2: Update OpenAPI Spec

**File:** `internal/api/file/gen/api.yaml`

### 2a: Change POST /file request body to multipart/form-data

Replace `application/json` + `FileUploadRequest` with:

```yaml
requestBody:
  description: The file to upload.
  required: true
  content:
    multipart/form-data:
      schema:
        type: object
        properties:
          name:
            type: string
            description: The name of the file in the Object Store.
            example: 'nginx.conf'
          content_type:
            type: string
            description: >
              How the file should be treated during deploy. "raw" writes bytes
              as-is; "template" renders with Go text/template and agent facts.
            default: raw
            enum:
              - raw
              - template
          file:
            type: string
            format: binary
            description: The file content.
        required:
          - name
          - file
```

### 2b: Add `content_type` to response schemas

Add `content_type` string field to `FileUploadResponse`, `FileInfo`, and
`FileInfoResponse`. Add to their `required` arrays.

### 2c: Regenerate

```bash
go generate ./internal/api/file/gen/...
```

---

## Step 3: Rewrite Upload Handler

**File:** `internal/api/file/file_upload.go`

Replace JSON-based handler with multipart streaming:

1. Extract form fields (name, content_type) from multipart body
2. Validate name manually (multipart fields don't use struct tags)
3. Open multipart file as `io.ReadSeeker`
4. First pass: compute SHA-256 via `io.Copy(hash, file)`
5. Idempotency check against existing digest
6. Second pass: `file.Seek(0, 0)` then stream to NATS via `Put(meta, file)`
7. Store content_type as `Osapi-Content-Type` NATS header on the object

If oapi-codegen strict-server doesn't parse multipart correctly, fall back to
custom Echo handler registered in `handler_file.go`.

---

## Step 4: Add `content_type` to Get/List Handlers

**Files:** `internal/api/file/file_get.go`, `internal/api/file/file_list.go`

Read `Osapi-Content-Type` from NATS object headers and include in responses.

---

## Step 5: Increase Object Store Bucket Size

**Files:** `configs/osapi.yaml`, `configs/osapi.local.yaml`

Change `max_bytes` from `104857600` (100 MiB) to `10737418240` (10 GiB).

---

## Step 6: Update Tests

**File:** `internal/api/file/file_upload_public_test.go`

- Rewrite `TestPostFile` for multipart request objects
- Rewrite `TestPostFileHTTP` to send `multipart/form-data`
- Rewrite `TestPostFileRBACHTTP` similarly
- Update file_get and file_list tests to assert `content_type`
- Add mock expectations for `Put` (streaming) instead of `PutBytes`

---

## Step 7: Update CLI

**File:** `cmd/client_file_upload.go`

- Add `--content-type` flag (default `raw`)
- Stream file from disk via `os.Open` instead of `os.ReadFile`
- Pass content_type to SDK `Upload` call
- Show `Content-Type` in output

---

## Step 8: Update SDK

**Files in `osapi-sdk`:**

- Copy updated `api.yaml` to SDK, regenerate with `redocly join` + `go generate`
- Add `ContentType` field to `FileUpload`, `FileItem`, `FileMetadata` types
- Change `Upload` method to accept `io.Reader` and `contentType` parameter
- Build multipart request body in SDK

---

## Verification

```bash
go generate ./internal/api/file/gen/...
go generate ./internal/api/file/mocks/...
go build ./...
go test ./internal/api/file/... -count=1 -v
just go::unit
just go::vet
```
