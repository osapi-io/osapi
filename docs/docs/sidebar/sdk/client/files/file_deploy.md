---
sidebar_position: 2
---

# NodeFile

File deploy operations on target hosts: deploy files from the Object Store to
agents, check the status and undeploy.

## Methods

| Method                                                        | Returns                                  | Permission   | Description                                         |
| ------------------------------------------------------------- | ---------------------------------------- | ------------ | --------------------------------------------------- |
| `c.NodeFile.Status(ctx, hostname, path)`                      | `Response[Collection[FileStatus]]`       | `file:read`  | Show whether a deployed file holds what osapi wrote |
| `c.NodeFile.Update(ctx, hostname, input NodeFileUpdateInput)` | `Response[Collection[FileUpdateResult]]` | `file:write` | Deploy an object of the object store to a path      |
| `c.NodeFile.Delete(ctx, hostname, input NodeFileDeleteInput)` | `Response[Collection[FileDeleteResult]]` | `file:write` | Remove a file that osapi deployed                   |

## Types

The fields of each type are the properties of its schema in the API's spec.

### FileStatus

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Path`     | `string` | Absolute path of the file.                                                             |
| `SHA256`   | `string` | SHA-256 of the file on the host, when it is there.                                     |
| `State`    | `string` | In-sync, drifted, missing or unmanaged.                                                |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### NodeFileUpdateInput

| Field         | Type                | Description                                                                                                                                               |
| ------------- | ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ContentType` | `string`            | How the object is written: raw as it is, or template rendered with .Facts, .Vars and .Hostname; the object's own when empty. It is one of: raw, template. |
| `Group`       | `string`            | Group that owns the file; left as written when empty.                                                                                                     |
| `Mode`        | `string`            | Permissions in octal, such as 0644; when empty, a file keeps its mode and a new file gets 0644.                                                           |
| `ObjectName`  | `string`            | The object of the object store, from osapi client file list.                                                                                              |
| `Owner`       | `string`            | User that owns the file; left as written when empty.                                                                                                      |
| `Path`        | `string`            | Absolute path of the file on the host.                                                                                                                    |
| `Vars`        | `map[string]string` | Variables of a template, as .Vars, each as key=value.                                                                                                     |

### FileUpdateResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Path`     | `string` | Absolute path of the file.                                                             |
| `SHA256`   | `string` | SHA-256 of the content that the file holds.                                            |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### NodeFileDeleteInput

| Field  | Type     | Description                            |
| ------ | -------- | -------------------------------------- |
| `Path` | `string` | Absolute path of the file on the host. |

### FileDeleteResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Path`     | `string` | Absolute path of the file.                                                             |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Deploy a raw file to a specific host
resp, err := c.NodeFile.Update(ctx, "web-01", client.NodeFileUpdateInput{
    ObjectName:  "nginx.conf",
    Path:        "/etc/nginx/nginx.conf",
    ContentType: "raw",
    Mode:        "0644",
})

// Deploy a template with variables
resp, err := c.NodeFile.Update(ctx, "_all", client.NodeFileUpdateInput{
    ObjectName:  "app.conf.tmpl",
    Path:        "/etc/app/config.yaml",
    ContentType: "template",
    Vars: map[string]string{
        "port":  "8080",
        "debug": "false",
    },
})

// Check file status on a host
resp, err := c.NodeFile.Status(ctx, "web-01", "/etc/nginx/nginx.conf")
```

## Example

Refer to
[`examples/sdk/client/file_deploy.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/file_deploy.go)
for a complete working example. See also
[`examples/sdk/client/file.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/file.go)
for Object Store operations (upload, list, get, delete).

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method. The methods run on each Linux host and on macOS.
