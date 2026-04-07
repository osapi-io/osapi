---
sidebar_position: 1
---

# Command

Command execution operations -- direct exec and shell-interpreted commands.

## Methods

| Method            | Description                                 |
| ----------------- | ------------------------------------------- |
| `Exec(ctx, req)`  | Execute a command directly (no shell)       |
| `Shell(ctx, req)` | Execute via `/bin/sh -c` (pipes, redirects) |

## Request Types

| Type           | Fields                           |
| -------------- | -------------------------------- |
| `ExecRequest`  | Target, Command, Args (optional) |
| `ShellRequest` | Target, Command                  |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Execute a command directly
resp, err := c.Command.Exec(ctx, client.ExecRequest{
    Command: "apt",
    Args:    []string{"install", "-y", "nginx"},
    Target:  "_all",
})

// Execute a shell command (pipes, redirection)
resp, err := c.Command.Shell(ctx, client.ShellRequest{
    Command: "ps aux | grep nginx",
    Target:  "_any",
})
```

## Example

See
[`examples/sdk/client/command.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/command.go)
for a complete working example.

## Permissions

Requires `command:execute` permission.
