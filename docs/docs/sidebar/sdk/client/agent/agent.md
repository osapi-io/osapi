---
sidebar_position: 1
---

# Agent

Agent discovery, details, and lifecycle management.

## Methods

| Method                               | Description                           |
| ------------------------------------ | ------------------------------------- |
| `List(ctx)`                          | Retrieve all active agents            |
| `Get(ctx, hostname)`                 | Get detailed agent info by hostname   |
| `Drain(ctx, hostname)`               | Drain agent (stop accepting jobs)     |
| `Undrain(ctx, hostname)`             | Undrain agent (resume accepting jobs) |
| `ListPending(ctx)`                   | List agents awaiting PKI enrollment   |
| `Accept(ctx, hostname, fingerprint)` | Accept a pending enrollment request   |
| `Reject(ctx, hostname)`              | Reject a pending enrollment request   |
| `Delete(ctx, hostname, machineID)`   | Delete an enrolled agent's stored key |

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

### Enrollment

When PKI is enabled, agents submit enrollment requests that must be accepted
before they join the fleet.

```go
// List pending enrollment requests
pending, err := c.Agent.ListPending(ctx)
for _, a := range pending.Data.Items {
    fmt.Printf("%s  fingerprint=%s\n",
        a.Hostname, a.Fingerprint)
}

// Accept by hostname
err := c.Agent.Accept(ctx, "web-01", "")

// Accept by fingerprint
err := c.Agent.Accept(ctx, "web-01", "SHA256:ab12cd34...")

// Reject
err := c.Agent.Reject(ctx, "web-01")

// Delete an enrolled agent. Pass a machine ID when two agents share the
// hostname; otherwise the call returns a ConflictError naming both.
err := c.Agent.Delete(ctx, "web-01", "")
```

## Examples

Refer to
[`examples/sdk/client/agent.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/agent.go)
for agent discovery and facts.

Refer to
[`examples/sdk/client/enrollment.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/enrollment.go)
for PKI enrollment operations.

## Enrollment Types

### `PendingAgent`

| Field         | Type        | Description                   |
| ------------- | ----------- | ----------------------------- |
| `MachineID`   | `string`    | Machine ID of the agent       |
| `Hostname`    | `string`    | Agent hostname                |
| `Fingerprint` | `string`    | SHA256 public key fingerprint |
| `RequestedAt` | `time.Time` | When enrollment was requested |

### `PendingAgentList`

| Field        | Type             | Description                    |
| ------------ | ---------------- | ------------------------------ |
| `Items`      | `[]PendingAgent` | List of pending agents         |
| `TotalItems` | `int`            | Total number of pending agents |

## Permissions

Requires `agent:read` for List, Get, and ListPending. Drain, Undrain, Accept,
Reject, and Delete require `agent:write`. Drain, Undrain, Accept, Reject and
Delete return only an error: the API answers 204.
