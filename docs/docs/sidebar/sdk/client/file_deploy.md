---
sidebar_position: 8
---

# FileDeploy

File deployment operations on target hosts -- deploy files from the Object Store
to agents, check status, and undeploy.

## Methods

| Method                      | Description                         |
| --------------------------- | ----------------------------------- |
| `Deploy(ctx, opts)`         | Deploy file to agent with SHA check |
| `Undeploy(ctx, opts)`       | Remove a deployed file              |
| `Status(ctx, target, path)` | Check deployed file status          |

## FileDeployOpts

| Field         | Type           | Required | Description                          |
| ------------- | -------------- | -------- | ------------------------------------ |
| `ObjectName`  | string         | Yes      | Name of the file in Object Store     |
| `Path`        | string         | Yes      | Destination path on the target host  |
| `ContentType` | string         | Yes      | `"raw"` or `"template"`             |
| `Mode`        | string         | No       | File permission mode (e.g. `"0644"`) |
| `Owner`       | string         | No       | File owner user                      |
| `Group`       | string         | No       | File owner group                     |
| `Vars`        | map[string]any | No       | Template variables for `"template"`  |
| `Target`      | string         | Yes      | Host target (see Targeting below)    |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Deploy a raw file to a specific host
resp, err := c.FileDeploy.Deploy(ctx, client.FileDeployOpts{
    ObjectName:  "nginx.conf",
    Path:        "/etc/nginx/nginx.conf",
    ContentType: "raw",
    Mode:        "0644",
    Target:      "web-01",
})

// Deploy a template with variables
resp, err := c.FileDeploy.Deploy(ctx, client.FileDeployOpts{
    ObjectName:  "app.conf.tmpl",
    Path:        "/etc/app/config.yaml",
    ContentType: "template",
    Vars: map[string]any{
        "port":  8080,
        "debug": false,
    },
    Target: "_all",
})

// Check file status on a host
resp, err := c.FileDeploy.Status(ctx, "web-01", "/etc/nginx/nginx.conf")
```

## Example

See
[`examples/sdk/client/file_deploy.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/file_deploy.go)
for a complete working example. See also
[`examples/sdk/client/file.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/file.go)
for Object Store operations (upload, list, get, delete).

## Permissions

Deploy and undeploy require `file:write`. Status requires `file:read`.
