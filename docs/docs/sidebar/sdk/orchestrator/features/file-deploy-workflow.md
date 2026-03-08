---
sidebar_position: 9
---

# File Deployment

Orchestrate a full file deployment workflow: upload a template to the Object
Store, deploy it to agents with template rendering, then verify status.

## Usage

```go
// Step 1: Upload the template file.
upload := plan.TaskFunc("upload-template",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        resp, err := c.File.Upload(ctx, "app.conf.tmpl", "template",
            bytes.NewReader(tmpl), client.WithForce())
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{Changed: resp.Data.Changed}, nil
    },
)

// Step 2: Deploy to all agents.
deploy := plan.Task("deploy-config", &orchestrator.Op{
    Operation: "file.deploy.execute",
    Target:    "_all",
    Params: map[string]any{
        "object_name":  "app.conf.tmpl",
        "path":         "/etc/app/config.yaml",
        "content_type": "template",
        "vars":         map[string]any{"port": 8080},
    },
})
deploy.DependsOn(upload)

// Step 3: Verify the deployed file.
verify := plan.Task("verify-status", &orchestrator.Op{
    Operation: "file.status.get",
    Target:    "_all",
    Params:    map[string]any{"path": "/etc/app/config.yaml"},
})
verify.DependsOn(deploy)
```

## Example

See
[`examples/sdk/orchestrator/file-deploy.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/file-deploy.go)
for a complete working example.
