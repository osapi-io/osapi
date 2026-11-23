---
sidebar_position: 1
---

# Command

Command execution operations: direct exec, and commands through a shell.

## Methods

| Method                                                    | Returns                               | Permission        | Description                           |
| --------------------------------------------------------- | ------------------------------------- | ----------------- | ------------------------------------- |
| `c.Command.Exec(ctx, hostname, input CommandExecInput)`   | `Response[Collection[CommandResult]]` | `command:execute` | Run a program with its arguments      |
| `c.Command.Shell(ctx, hostname, input CommandShellInput)` | `Response[Collection[CommandResult]]` | `command:shell`   | Run a command line through /bin/sh -c |

## Types

The fields of each type are the properties of its schema in the API's spec.

### CommandExecInput

| Field     | Type       | Description                                                       |
| --------- | ---------- | ----------------------------------------------------------------- |
| `Args`    | `[]string` | Its arguments, each one word, such as -la /tmp.                   |
| `Command` | `string`   | The program, a name on the agent's PATH or a path, such as ls.    |
| `Cwd`     | `string`   | The directory to run it in; the agent's own when empty.           |
| `Timeout` | `int`      | Seconds to let it run before it is stopped, up to 300; 30 when 0. |

### CommandResult

| Field        | Type     | Description                                                                            |
| ------------ | -------- | -------------------------------------------------------------------------------------- |
| `Changed`    | `bool`   | Whether the host changed: true, since a command can change anything.                   |
| `DurationMS` | `int64`  | How long it ran, in milliseconds.                                                      |
| `Error`      | `string` | Error message if the agent failed.                                                     |
| `ExitCode`   | `int`    | Its exit code.                                                                         |
| `Hostname`   | `string` | Hostname of the agent that produced this result.                                       |
| `Status`     | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Stderr`     | `string` | What it wrote to standard error.                                                       |
| `Stdout`     | `string` | What it wrote to standard output.                                                      |

### CommandShellInput

| Field     | Type     | Description                                                                |
| --------- | -------- | -------------------------------------------------------------------------- |
| `Command` | `string` | The command line, which /bin/sh -c reads, such as ls -la /tmp \| grep log. |
| `Cwd`     | `string` | The directory to run it in; the agent's own when empty.                    |
| `Timeout` | `int`    | Seconds to let it run before it is stopped, up to 300; 30 when 0.          |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Execute a command directly
resp, err := c.Command.Exec(ctx, "_all", client.CommandExecInput{
    Command: "apt",
    Args:    []string{"install", "-y", "nginx"},
})

// Execute a shell command (pipes, redirection)
resp, err := c.Command.Shell(ctx, "_any", client.CommandShellInput{
    Command: "ps aux | grep nginx",
})
```

## Example

Refer to
[`examples/sdk/client/command.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/command.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
