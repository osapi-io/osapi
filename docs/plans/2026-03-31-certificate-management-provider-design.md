# Certificate Management Provider Design

## Overview

Add CA certificate management to OSAPI. Deploy custom CA
certificates to the system trust store alongside default system
CAs, and remove them. Uses `file.Deployer` for SHA-tracked
deployment and `update-ca-certificates` to rebuild the trust
bundle. Read-only listing of both system and custom CAs.

## Architecture

Meta provider at `internal/provider/node/certificate/`.

- **Category**: `node`
- **Path prefix**: `/node/{hostname}/certificate/ca`
- **Permissions**: `certificate:read`, `certificate:write`
- **Provider type**: meta (file.Deployer + exec.Manager)

## Provider Interface

```go
type Provider interface {
    List(ctx context.Context) ([]Entry, error)
    Create(ctx context.Context, entry Entry) (*CreateResult, error)
    Update(ctx context.Context, entry Entry) (*UpdateResult, error)
    Delete(ctx context.Context, name string) (*DeleteResult, error)
}
```

## Data Types

```go
type Entry struct {
    Name   string `json:"name"`
    Source string `json:"source"` // "system" or "custom"
    Object string `json:"object,omitempty"`
}

type CreateResult struct {
    Changed bool   `json:"changed"`
    Name    string `json:"name"`
}

type UpdateResult struct {
    Changed bool   `json:"changed"`
    Name    string `json:"name"`
}

type DeleteResult struct {
    Changed bool   `json:"changed"`
    Name    string `json:"name"`
}
```

## Debian Implementation

Custom CA certs are deployed to
`/usr/local/share/ca-certificates/osapi-{name}.crt` via
`file.Deployer`. After every create, update, or delete, the
provider runs `update-ca-certificates` to rebuild the system
trust bundle.

- **List**: Walk `/usr/share/ca-certificates/` for system CAs
  (strip path prefix and `.crt` extension for name). Query file
  state KV for entries with `osapi-` prefix for custom CAs.
  Return both with `source` field.
- **Create**: Deploy PEM from Object Store to
  `/usr/local/share/ca-certificates/osapi-{name}.crt` via
  `file.Deployer` with mode `0644`. Store metadata
  `{"source": "custom"}` in FileState. Run
  `update-ca-certificates`.
- **Update**: Same as create but for an existing entry. The
  `file.Deployer` compares SHA — if content unchanged, returns
  `changed: false` and skips `update-ca-certificates`.
- **Delete**: Undeploy via `file.Deployer`, run
  `update-ca-certificates`.

## Platform Implementations

| Platform | Implementation                        |
| -------- | ------------------------------------- |
| Debian   | file.Deployer + update-ca-certificates |
| Darwin   | ErrUnsupported                        |
| Linux    | ErrUnsupported                        |

## Container Behavior

No container check — CA cert management works in Docker
containers. `update-ca-certificates` is available and the trust
store is writable.

## API Endpoints

| Method   | Path                                       | Permission          | Description            |
| -------- | ------------------------------------------ | ------------------- | ---------------------- |
| `GET`    | `/node/{hostname}/certificate/ca`          | `certificate:read`  | List CA certs          |
| `POST`   | `/node/{hostname}/certificate/ca`          | `certificate:write` | Add custom CA cert     |
| `PUT`    | `/node/{hostname}/certificate/ca/{name}`   | `certificate:write` | Update custom CA cert  |
| `DELETE` | `/node/{hostname}/certificate/ca/{name}`   | `certificate:write` | Remove custom CA cert  |

All endpoints support broadcast targeting.

### POST/PUT Request Body

```json
{
  "name": "internal-corp-ca",
  "object": "corp-ca.pem"
}
```

`object` references an existing Object Store upload containing
the PEM-encoded CA certificate. For PUT, `name` comes from the
path parameter.

### Response Shape (List)

```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "certificates": [
      {"name": "DigiCert_Global_Root_G2", "source": "system"},
      {"name": "internal-corp-ca", "source": "custom"}
    ]
  }]
}
```

### Response Shape (Create/Update/Delete)

```json
{
  "job_id": "...",
  "results": [{
    "hostname": "web-01",
    "status": "ok",
    "name": "internal-corp-ca",
    "changed": true
  }]
}
```

## SDK

```go
client.Certificate.List(ctx, host)
client.Certificate.Create(ctx, host, opts)
client.Certificate.Update(ctx, host, name, opts)
client.Certificate.Delete(ctx, host, name)
```

`CertificateCreateOpts` / `CertificateUpdateOpts` with `Name`
and `Object` fields.

## Permissions

- `certificate:read` — list CA certificates. Added to admin,
  write, and read roles.
- `certificate:write` — create, update, delete custom CA
  certificates. Added to admin and write roles.
