---
sidebar_position: 3
---

# Process

The `Process` service has methods that list processes, get process details and
send signals to processes on target hosts. Use `client.Process.List()`,
`client.Process.Get()` and `client.Process.Signal()`.

## Methods

| Method                                                      | Returns                               | Permission        | Description                    |
| ----------------------------------------------------------- | ------------------------------------- | ----------------- | ------------------------------ |
| `c.Process.List(ctx, hostname)`                             | `Response[Collection[ProcessEntry]]`  | `process:read`    | List the processes of the host |
| `c.Process.Get(ctx, hostname, pid)`                         | `Response[Collection[ProcessEntry]]`  | `process:read`    | Get one process                |
| `c.Process.Signal(ctx, hostname, input ProcessSignalInput)` | `Response[Collection[ProcessResult]]` | `process:execute` | Send a signal to a process     |

## Types

The fields of each type are the properties of its schema in the API's spec.

### ProcessEntry

| Field        | Type      | Description                                                                            |
| ------------ | --------- | -------------------------------------------------------------------------------------- |
| `Command`    | `string`  | Its full command line.                                                                 |
| `CPUPercent` | `float64` | Its CPU time as a percent of the time since it started.                                |
| `Error`      | `string`  | Error message if the agent failed.                                                     |
| `Hostname`   | `string`  | Hostname of the agent that produced this result.                                       |
| `MemPercent` | `float64` | Its resident memory as a percent of the host's memory.                                 |
| `MemRSS`     | `int64`   | Its resident memory in bytes.                                                          |
| `Name`       | `string`  | The name of the process.                                                               |
| `PID`        | `int`     | The process ID.                                                                        |
| `StartTime`  | `string`  | When it started, in RFC 3339.                                                          |
| `State`      | `string`  | Its state, such as running or sleep.                                                   |
| `Status`     | `string`  | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `User`       | `string`  | The user that runs it.                                                                 |

### ProcessSignalInput

| Field    | Type     | Description                                                                                     |
| -------- | -------- | ----------------------------------------------------------------------------------------------- |
| `PID`    | `int`    | The pid of the path: the process ID, such as 1234.                                              |
| `Signal` | `string` | The signal: TERM, KILL, HUP, INT, USR1 or USR2. It is one of: TERM, KILL, HUP, INT, USR1, USR2. |

### ProcessResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the signal was sent.                                                           |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `PID`      | `int`    | The process ID.                                                                        |
| `Signal`   | `string` | The signal that was sent.                                                              |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

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
sigResp, err := c.Process.Signal(ctx, "web-01", client.ProcessSignalInput{
    PID: 1234,
    Signal: "TERM",
})
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

The Permission column of [Methods](#methods) gives the permission of each
method.

The process behavior operates on each Linux family and on macOS. The `operator`
and `admin` roles get `process:execute`.
