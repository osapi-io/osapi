---
sidebar_position: 6
---

# file.upload

Upload file content to the OSAPI Object Store. Returns the object name that can
be referenced in subsequent `file.deploy.execute` operations.

## Usage

```go
task := plan.TaskFunc("upload-config",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.File.Upload(
            ctx,
            "nginx.conf",
            "application/octet-stream",
            bytes.NewReader(configBytes),
        )
        if err != nil {
            return nil, err
        }

        return &orchestrator.Result{
            Changed: true,
            Data:    orchestrator.StructToMap(resp.Data),
        }, nil
    },
)
```

## Parameters

| Param          | Type      | Required | Description                     |
| -------------- | --------- | -------- | ------------------------------- |
| `name`         | string    | Yes      | Object name in the Object Store |
| `content_type` | string    | Yes      | MIME type of the content        |
| `content`      | io.Reader | Yes      | File content to upload          |

## Target

Not applicable. Upload is a server-side operation that does not target an agent.

## Idempotency

**Idempotent.** Uploading the same content with the same name overwrites the
existing object.

## Permissions

Requires `file:write` permission.

## Example

See
[`examples/sdk/orchestrator/operations/file-upload.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/file-upload.go)
for a complete working example.
