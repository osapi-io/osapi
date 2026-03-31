---
sidebar_position: 2
---

# Agent

Agent discovery, details, and lifecycle management.

## Methods

| Method                  | Description                         |
| ----------------------- | ----------------------------------- |
| `List(ctx)`             | Retrieve all active agents          |
| `Get(ctx, hostname)`    | Get detailed agent info by hostname |
| `Drain(ctx, hostname)`  | Drain agent (stop accepting jobs)   |
| `Undrain(ctx, hostname)`| Undrain agent (resume accepting jobs) |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all agents
resp, err := c.Agent.List(ctx)
for _, a := range resp.Data.Agents {
    fmt.Printf("%s  status=%s\n", a.Hostname, a.Status)
}

// Get specific agent details
resp, err := c.Agent.Get(ctx, "web-01")

// Drain an agent (stop new jobs)
resp, err := c.Agent.Drain(ctx, "web-01")

// Undrain an agent (resume jobs)
resp, err := c.Agent.Undrain(ctx, "web-01")
```

## Example

See
[`examples/sdk/client/agent.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/agent.go)
for a complete working example.

## Permissions

Requires `agent:read` for List and Get. Drain and Undrain require
`agent:write`.
