---
sidebar_position: 4
---

# FileService

File management operations for the OSAPI Object Store. Upload, list, inspect,
and delete files that can be deployed to agents via `Node.FileDeploy`.

## Methods

### Object Store

| Method                          | Description                                     |
| ------------------------------- | ----------------------------------------------- |
| `Upload(ctx, name, ct, r, ...)` | Upload file content to Object Store             |
| `Changed(ctx, name, r)`         | Check if local content differs from stored file |
| `List(ctx)`                     | List all stored files                           |
| `Get(ctx, name)`                | Get file metadata by name                       |
| `Delete(ctx, name)`             | Delete a file from Object Store                 |

### Node File Operations

File deploy and status methods live on `NodeService` because they target a
specific host:

| Method                          | Description                         |
| ------------------------------- | ----------------------------------- |
| `FileDeploy(ctx, opts)`         | Deploy file to agent with SHA check |
| `FileStatus(ctx, target, path)` | Check deployed file status          |

## FileDeployOpts

| Field         | Type           | Required | Description                          |
| ------------- | -------------- | -------- | ------------------------------------ |
| `ObjectName`  | string         | Yes      | Name of the file in Object Store     |
| `Path`        | string         | Yes      | Destination path on the target host  |
| `ContentType` | string         | Yes      | `"raw"` or `"template"`              |
| `Mode`        | string         | No       | File permission mode (e.g. `"0644"`) |
| `Owner`       | string         | No       | File owner user                      |
| `Group`       | string         | No       | File owner group                     |
| `Vars`        | map[string]any | No       | Template variables for `"template"`  |
| `Target`      | string         | Yes      | Host target (see Targeting below)    |

## Upload Options

| Option        | Description                                             |
| ------------- | ------------------------------------------------------- |
| `WithForce()` | Bypass SDK-side and server-side SHA check; always write |

## Usage

```go
// Upload a raw file.
resp, err := client.File.Upload(
    ctx, "nginx.conf", "raw", bytes.NewReader(data),
)

// Force upload — skip SHA-256 check, always write.
resp, err := client.File.Upload(
    ctx, "nginx.conf", "raw", bytes.NewReader(data),
    client.WithForce(),
)

// Check if content differs without uploading.
chk, err := client.File.Changed(
    ctx, "nginx.conf", bytes.NewReader(data),
)
fmt.Println(chk.Data.Changed) // true if content differs

// List all files.
resp, err := client.File.List(ctx)

// Get file metadata.
resp, err := client.File.Get(ctx, "nginx.conf")

// Delete a file.
resp, err := client.File.Delete(ctx, "nginx.conf")

// Deploy a raw file to a specific host.
resp, err := client.Node.FileDeploy(ctx, client.FileDeployOpts{
    ObjectName:  "nginx.conf",
    Path:        "/etc/nginx/nginx.conf",
    ContentType: "raw",
    Mode:        "0644",
    Owner:       "root",
    Group:       "root",
    Target:      "web-01",
})

// Deploy a template file with variables.
resp, err := client.Node.FileDeploy(ctx, client.FileDeployOpts{
    ObjectName:  "app.conf.tmpl",
    Path:        "/etc/app/config.yaml",
    ContentType: "template",
    Vars: map[string]any{
        "port":  8080,
        "debug": false,
    },
    Target: "_all",
})

// Check file status on a host.
resp, err := client.Node.FileStatus(
    ctx, "web-01", "/etc/nginx/nginx.conf",
)
```

## Targeting

`FileDeploy` and `FileStatus` accept any valid target: `_any`, `_all`, a
hostname, or a label selector (`key:value`).

Object Store operations (`Upload`, `List`, `Get`, `Delete`) are server-side and
do not use targeting.

## Change Detection

`Upload` computes a SHA-256 of the file content locally before uploading. If the
hash matches the stored file, the upload is skipped and `Changed: false` is
returned. Use `WithForce()` to bypass this check.

`Changed` performs the same SHA-256 comparison without uploading. It returns
`Changed: true` when the file does not exist or the content differs.

## Idempotency

`FileDeploy` compares the SHA-256 of the Object Store content against the
deployed file. If the hashes match, the file is not rewritten and the response
reports `Changed: false`.

## Example

See
[`examples/sdk/client/file.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/file.go)
for a complete working example.

## Permissions

Object Store operations require `file:read` (list, get) or `file:write` (upload,
delete). Deploy requires `file:write`. Status requires `file:read`.
