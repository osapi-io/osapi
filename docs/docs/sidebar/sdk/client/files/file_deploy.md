---
sidebar_position: 2
---

# FileDeploy

File deploy operations on target hosts: deploy files from the Object Store to
agents, check the status and undeploy.

## Methods

| Method                      | Description                         |
| --------------------------- | ----------------------------------- |
| `Deploy(ctx, opts)`         | Deploy file to agent with SHA check |
| `Undeploy(ctx, opts)`       | Remove a deployed file              |
| `Status(ctx, target, path)` | Check deployed file status          |

## FileDeployOpts

| Field         | Type              | Required | Description                                                    |
| ------------- | ----------------- | -------- | -------------------------------------------------------------- |
| `ObjectName`  | string            | Yes      | Name of the file in Object Store                               |
| `Path`        | string            | Yes      | Absolute destination path on the target host                   |
| `ContentType` | `ContentType`     | No       | `ContentRaw` or `ContentTemplate`; the object's own when empty |
| `Mode`        | string            | No       | File permission mode (e.g. `"0644"`)                           |
| `Owner`       | string            | No       | File owner user                                                |
| `Group`       | string            | No       | File owner group                                               |
| `Vars`        | map[string]string | No       | Template variables, as `.Vars`                                 |
| `Target`      | string            | Yes      | Host target (see Targeting below)                              |

`Status` returns a `FileStatusResult` for each host. Its `State` is `in-sync`,
`drifted` or `missing`. Its `Status` is the status of the job on the host: `ok`,
`failed` or `skipped`.

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

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
    Vars: map[string]string{
        "port":  "8080",
        "debug": "false",
    },
    Target: "_all",
})

// Check file status on a host
resp, err := c.FileDeploy.Status(ctx, "web-01", "/etc/nginx/nginx.conf")
```

## Example

Refer to
[`examples/sdk/client/file_deploy.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/file_deploy.go)
for a complete working example. See also
[`examples/sdk/client/file.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/file.go)
for Object Store operations (upload, list, get, delete).

## Permissions

Deploy and undeploy are actions, and require `file:execute`. Status requires
`file:read`. They run on each Linux host. A macOS host reports `skipped`.
