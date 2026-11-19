---
sidebar_position: 1
---

# Agent

Agent discovery, details, and lifecycle management.

## Methods

| Method                   | Description                           |
| ------------------------ | ------------------------------------- |
| `List(ctx)`              | Retrieve all active agents            |
| `Get(ctx, hostname)`     | Get detailed agent info by hostname   |
| `Drain(ctx, hostname)`   | Drain agent (stop accepting jobs)     |
| `Undrain(ctx, hostname)` | Undrain agent (resume accepting jobs) |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all agents
resp, err := c.Agent.List(ctx)
for _, a := range resp.Data.Items {
    fmt.Printf("%s  status=%s\n", a.Hostname, a.Status)
}

// Get specific agent details
resp, err := c.Agent.Get(ctx, "web-01")

// Drain an agent (stop new jobs)
err := c.Agent.Drain(ctx, "web-01")

// Undrain an agent (resume jobs)
err := c.Agent.Undrain(ctx, "web-01")
```

## Examples

Refer to
[`examples/sdk/client/agent.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/agent.go)
for agent discovery and facts.

## Permissions

List and Get need `agent:read`. Drain and Undrain need `agent:write`, and return
only an error: the API answers 204. To admit, reject or revoke a machine, use
the [Admission](admission.md) service.
