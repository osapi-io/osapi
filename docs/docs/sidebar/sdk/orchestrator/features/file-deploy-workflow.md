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
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.File.Upload(
            ctx, "app.conf.tmpl", "template",
            bytes.NewReader(tmpl), client.WithForce(),
        )
        if err != nil {
            return nil, err
        }
        return &orchestrator.Result{
            Changed: resp.Data.Changed,
        }, nil
    },
)

// Step 2: Deploy to all agents.
deploy := plan.TaskFunc("deploy-config",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.FileDeploy(ctx, client.FileDeployOpts{
            Target:      "_all",
            ObjectName:  "app.conf.tmpl",
            Path:        "/etc/app/config.yaml",
            ContentType: "template",
            Vars:        map[string]any{"port": 8080},
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
deploy.DependsOn(upload)

// Step 3: Verify the deployed file.
verify := plan.TaskFunc("verify-status",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.FileStatus(
            ctx, "_all", "/etc/app/config.yaml",
        )
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
verify.DependsOn(deploy)
```

## Example

See
[`examples/sdk/orchestrator/file-deploy.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/file-deploy.go)
for a complete working example.
