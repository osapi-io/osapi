---
sidebar_position: 3
---

# Process

The `Process` service has methods that list processes, get process details and
send signals to processes on target hosts. Use `client.Process.List()`,
`client.Process.Get()` and `client.Process.Signal()`.

## Methods

| Method                             | Description                              |
| ---------------------------------- | ---------------------------------------- |
| `List(ctx, hostname)`              | List all running processes on the target |
| `Get(ctx, hostname, pid)`          | Get information about a process by PID   |
| `Signal(ctx, hostname, pid, opts)` | Send a signal to a process by PID        |

## Request Types

| Type                | Fields                             |
| ------------------- | ---------------------------------- |
| `ProcessSignalOpts` | `Signal` (string, e.g. TERM, KILL) |

## Result Types

| Type                  | Fields                                                                                                                      |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `ProcessInfoResult`   | `Hostname`, `Status`, `PID`, `Name`, `User`, `State`, `CPUPercent`, `MemPercent`, `MemRSS`, `Command`, `StartTime`, `Error` |
| `ProcessSignalResult` | `Hostname`, `Status`, `PID`, `Signal`, `Changed`, `Error`                                                                   |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List the processes. Each row is one process of one host.
resp, err := c.Process.List(ctx, "web-01")
for _, p := range resp.Data.Results {
    fmt.Printf("%s PID=%d name=%s cpu=%.1f%%\n",
        p.Hostname, p.PID, p.Name, p.CPUPercent)
}

// Get one process
resp, err = c.Process.Get(ctx, "web-01", 1234)
for _, p := range resp.Data.Results {
    fmt.Printf("PID=%d user=%s state=%s\n",
        p.PID, p.User, p.State)
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

- [`examples/sdk/client/process.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/process.go)

## Permissions

| Operation | Permission        |
| --------- | ----------------- |
| List, Get | `process:read`    |
| Signal    | `process:execute` |

The process behavior operates on each Linux family and on macOS. The `operator`
and `admin` roles get `process:execute`.
