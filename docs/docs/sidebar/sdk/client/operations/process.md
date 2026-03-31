---
sidebar_position: 3
---

# Process

The `Process` service provides methods for listing processes, getting process
details, and sending signals to processes on target hosts. Access via
`client.Process.List()`, `client.Process.Get()`, and
`client.Process.Signal()`.

## Methods

| Method                               | Description                                |
| ------------------------------------ | ------------------------------------------ |
| `List(ctx, hostname)`                | List all running processes on the target   |
| `Get(ctx, hostname, pid)`            | Get information about a process by PID     |
| `Signal(ctx, hostname, pid, opts)`   | Send a signal to a process by PID          |

## Request Types

| Type                | Fields                             |
| ------------------- | ---------------------------------- |
| `ProcessSignalOpts` | `Signal` (string, e.g. TERM, KILL) |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all processes
resp, err := c.Process.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    for _, p := range r.Processes {
        fmt.Printf("PID=%d name=%s cpu=%.1f%%\n",
            p.PID, p.Name, p.CPUPercent)
    }
}

// Get a specific process
resp, err := c.Process.Get(ctx, "web-01", 1234)
for _, r := range resp.Data.Results {
    for _, p := range r.Processes {
        fmt.Printf("PID=%d user=%s state=%s\n",
            p.PID, p.User, p.State)
    }
}

// Send TERM signal to a process
sigResp, err := c.Process.Signal(ctx, "web-01", 1234,
    client.ProcessSignalOpts{Signal: "TERM"})
for _, r := range sigResp.Data.Results {
    fmt.Printf("PID=%d signal=%s changed=%v\n",
        r.PID, r.Signal, r.Changed)
}

// Broadcast process list to all hosts
resp, err := c.Process.List(ctx, "_all")
```

## Example

- [`examples/sdk/client/process.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/process.go)

## Permissions

| Operation    | Permission        |
| ------------ | ----------------- |
| List, Get    | `process:read`    |
| Signal       | `process:execute` |

Process management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
