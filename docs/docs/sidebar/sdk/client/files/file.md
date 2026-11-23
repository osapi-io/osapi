---
sidebar_position: 1
---

# File

File management operations for the OSAPI Object Store. Upload, list, examine and
delete files that you can deploy to agents with
[`NodeFile.Update`](file_deploy.md).

## Methods

| Method                                                  | Returns                              | Permission   | Description                                                               |
| ------------------------------------------------------- | ------------------------------------ | ------------ | ------------------------------------------------------------------------- |
| `c.File.List(ctx)`                                      | `Response[FileListResponse]`         | `file:read`  | List stored files                                                         |
| `c.File.Stale(ctx)`                                     | `Response[StaleDeploymentsResponse]` | `file:read`  | List stale file deployments                                               |
| `c.File.Get(ctx, name)`                                 | `Response[FileInfoResponse]`         | `file:read`  | Get file metadata                                                         |
| `c.File.Delete(ctx, name)`                              | `error`                              | `file:write` | Delete a file                                                             |
| `c.File.Upload(ctx, name, contentType, file io.Reader)` | `Response[FileUploadResponse]`       | `file:write` | Upload a file; skips the upload when the stored file has the same SHA-256 |
| `c.File.Update(ctx, name, contentType, file io.Reader)` | `Response[FileUploadResponse]`       | `file:write` | Replace a file, or create it when the name is new                         |
| `c.File.GetChanged(ctx, name, file io.Reader)`          | `Response[FileChanged]`              | `file:read`  | Compare content with the stored file by SHA-256                           |

## Types

The fields of each type are the properties of its schema in the API's spec.

### FileListResponse

| Field        | Type         | Description            |
| ------------ | ------------ | ---------------------- |
| `Items`      | `[]FileInfo` | List of stored files.  |
| `TotalItems` | `int`        | Total number of files. |

### StaleDeploymentsResponse

| Field        | Type                | Description                        |
| ------------ | ------------------- | ---------------------------------- |
| `Items`      | `[]StaleDeployment` | The stale deployments.             |
| `TotalItems` | `int`               | Total number of stale deployments. |

### FileInfoResponse

| Field         | Type     | Description                                                     |
| ------------- | -------- | --------------------------------------------------------------- |
| `ContentType` | `string` | How the file should be treated during deploy (raw or template). |
| `Name`        | `string` | The name of the file.                                           |
| `SHA256`      | `string` | SHA-256 hash of the file content.                               |
| `Size`        | `int`    | File size in bytes.                                             |

## Usage

```go
// Upload a raw file.
resp, err := c.File.Upload(
    ctx, "nginx.conf", "raw", bytes.NewReader(data),
)

// Replace a file that exists, or create it.
resp, err := c.File.Update(
    ctx, "nginx.conf", "raw", bytes.NewReader(data),
)

// Check if content differs without uploading.
chk, err := c.File.GetChanged(
    ctx, "nginx.conf", bytes.NewReader(data),
)
fmt.Println(chk.Data.Changed) // true if content differs

// List all files.
resp, err := c.File.List(ctx)

// Get file metadata.
resp, err := c.File.Get(ctx, "nginx.conf")

// Delete a file.
err := c.File.Delete(ctx, "nginx.conf")

// Deploy a raw file to a specific host.
resp, err := c.NodeFile.Update(ctx, "web-01", client.NodeFileUpdateInput{
    ObjectName:  "nginx.conf",
    Path:        "/etc/nginx/nginx.conf",
    ContentType: "raw",
    Mode:        "0644",
    Owner:       "root",
    Group:       "root",
})

// Deploy a template file with variables.
resp, err := c.NodeFile.Update(ctx, "_all", client.NodeFileUpdateInput{
    ObjectName:  "app.conf.tmpl",
    Path:        "/etc/app/config.yaml",
    ContentType: "template",
    Vars: map[string]string{
        "port":  "8080",
        "debug": "false",
    },
})

// Check file status on a host.
resp, err := c.NodeFile.Status(
    ctx, "web-01", "/etc/nginx/nginx.conf",
)

// List stale deployments (objects updated since deploy).
stale, err := c.File.Stale(ctx)
for _, s := range stale.Data.Items {
    fmt.Printf("%s on %s is stale\n", s.ObjectName, s.Hostname)
}
```

## Targeting

`NodeFile.Update` and `NodeFile.Status` accept any valid target: `_any`, `_all`,
a hostname, or a label selector (`key:value`).

Object Store operations (`Upload`, `List`, `Get`, `Delete`) are server-side and
do not use targeting.

## Change Detection

`Upload` calculates the SHA-256 of the file content locally before it uploads
the file. If the hash matches the stored file, the upload is skipped and
`Changed: false` is returned. A file with other content is a `ConflictError`;
`Update` replaces it.

`GetChanged` does the same SHA-256 comparison, but does not upload. It returns
`Changed: true` when the file does not exist or the content differs.

## Idempotency

`NodeFile.Update` compares the SHA-256 of the Object Store content against the
deployed file. If the hashes match, the file is not rewritten and the response
reports `Changed: false`.

## Example

Refer to
[`examples/sdk/client/file.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/file.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
