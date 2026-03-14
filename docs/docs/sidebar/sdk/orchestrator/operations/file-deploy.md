---
sidebar_position: 4
---

# file.deploy.execute

Deploy a file from the Object Store to the target agent's filesystem. Supports
raw content and Go-template rendering with agent facts and custom variables.

## Usage

```go
task := plan.TaskFunc("deploy-config",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.FileDeploy(ctx, client.FileDeployOpts{
            Target:      "_all",
            ObjectName:  "nginx.conf",
            Path:        "/etc/nginx/nginx.conf",
            ContentType: "raw",
            Mode:        "0644",
            Owner:       "root",
            Group:       "root",
        })
        if err != nil {
            return nil, err
        }

        return &orchestrator.Result{
            JobID:   resp.Data.JobID,
            Changed: resp.Data.Changed,
            Data:    orchestrator.StructToMap(resp.Data),
        }, nil
    },
)
```

### Template Deployment

```go
task := plan.TaskFunc("deploy-template",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.FileDeploy(ctx, client.FileDeployOpts{
            Target:      "web-01",
            ObjectName:  "app.conf.tmpl",
            Path:        "/etc/app/config.yaml",
            ContentType: "template",
            Vars: map[string]any{
                "port":  8080,
                "debug": false,
            },
        })
        if err != nil {
            return nil, err
        }

        return &orchestrator.Result{
            JobID:   resp.Data.JobID,
            Changed: resp.Data.Changed,
            Data:    orchestrator.StructToMap(resp.Data),
        }, nil
    },
)
```

## Parameters

| Param          | Type           | Required | Description                              |
| -------------- | -------------- | -------- | ---------------------------------------- |
| `object_name`  | string         | Yes      | Name of the file in the Object Store     |
| `path`         | string         | Yes      | Destination path on the target host      |
| `content_type` | string         | Yes      | `"raw"` or `"template"`                  |
| `mode`         | string         | No       | File permission mode (e.g., `"0644"`)    |
| `owner`        | string         | No       | File owner user                          |
| `group`        | string         | No       | File owner group                         |
| `vars`         | map[string]any | No       | Template variables for `"template"` type |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Idempotent.** Compares the SHA-256 of the Object Store content against the
deployed file. Returns `Changed: true` only if the file was actually written.
Returns `Changed: false` if the hashes match.

## Permissions

Requires `file:write` permission.

## Example

See
[`examples/sdk/orchestrator/operations/file-deploy.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/file-deploy.go)
for a complete working example.
