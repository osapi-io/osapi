---
sidebar_position: 6
---

# Group

Group management on target hosts.

## Methods

| Method                                                  | Returns                              | Permission    | Description                 |
| ------------------------------------------------------- | ------------------------------------ | ------------- | --------------------------- |
| `c.Group.List(ctx, hostname)`                           | `Response[Collection[GroupAccount]]` | `group:read`  | List the groups of the host |
| `c.Group.Create(ctx, hostname, input GroupCreateInput)` | `Response[Collection[GroupResult]]`  | `group:write` | Add a group                 |
| `c.Group.Get(ctx, hostname, name)`                      | `Response[Collection[GroupAccount]]` | `group:read`  | Get one group               |
| `c.Group.Update(ctx, hostname, input GroupUpdateInput)` | `Response[Collection[GroupResult]]`  | `group:write` | Set the members of a group  |
| `c.Group.Delete(ctx, hostname, name)`                   | `Response[Collection[GroupResult]]`  | `group:write` | Remove a group              |

## Types

The fields of each type are the properties of its schema in the API's spec.

### GroupAccount

| Field      | Type       | Description                                                                            |
| ---------- | ---------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string`   | Error message if the agent failed.                                                     |
| `GID`      | `int`      | Its numeric group ID.                                                                  |
| `Hostname` | `string`   | Hostname of the agent that produced this result.                                       |
| `Members`  | `[]string` | The users it has as members, besides those whose primary group it is.                  |
| `Name`     | `string`   | The group.                                                                             |
| `Status`   | `string`   | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### GroupCreateInput

| Field    | Type     | Description                                                 |
| -------- | -------- | ----------------------------------------------------------- |
| `GID`    | `int`    | Its numeric group ID; the system assigns one when it is 0.  |
| `Name`   | `string` | The group to add, such as developers.                       |
| `System` | `bool`   | Add it as a system group, with an ID from the system range. |

### GroupResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The group.                                                                             |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### GroupUpdateInput

| Field     | Type       | Description                                                     |
| --------- | ---------- | --------------------------------------------------------------- |
| `Name`    | `string`   | The name of the path: the group to remove, such as developers.  |
| `Members` | `[]string` | Its members, which replace the members it has; none when empty. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all groups
resp, err := c.Group.List(ctx, "_any")
// Each result is one group on one host.
for _, r := range resp.Data.Results {
    fmt.Printf("%s: %s gid=%d\n", r.Hostname, r.Name, r.GID)
}

// Get a specific group
resp, err := c.Group.Get(ctx, "web-01", "docker")

// Create a group
resp, err := c.Group.Create(ctx, "web-01", client.GroupCreateInput{
    Name: "deploy",
})

// Update group members
resp, err := c.Group.Update(ctx, "web-01", client.GroupUpdateInput{
    Name: "deploy",
    Members: []string{"alice", "bob"},
})

// Delete a group
resp, err := c.Group.Delete(ctx, "web-01", "deploy")
```

## Example

Refer to
[`examples/sdk/client/group.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/group.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
