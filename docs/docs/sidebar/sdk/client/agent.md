---
sidebar_position: 2
---

# AgentService

Agent discovery and details.

## Methods

| Method               | Description                         |
| -------------------- | ----------------------------------- |
| `List(ctx)`          | Retrieve all active agents          |
| `Get(ctx, hostname)` | Get detailed agent info by hostname |

## Usage

```go
// List all agents
resp, err := client.Agent.List(ctx)

// Get specific agent details
resp, err := client.Agent.Get(ctx, "web-01")
```

## Example

See
[`examples/sdk/osapi/agent.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/osapi/agent.go)
for a complete working example.

## Permissions

Requires `agent:read` permission.
