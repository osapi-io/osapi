---
sidebar_position: 4
---

# Memory

Memory statistics query operations.

## Methods

| Method                        | Returns                             | Permission    | Description                      |
| ----------------------------- | ----------------------------------- | ------------- | -------------------------------- |
| `c.Memory.Get(ctx, hostname)` | `Response[Collection[MemoryEntry]]` | `memory:read` | Get the memory usage of the host |

## Types

The fields of each type are the properties of its schema in the API's spec.

### MemoryEntry

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Memory`   | `Usage`  | Memory is the property memory.                                                         |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get memory stats from all hosts
resp, err := c.Memory.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    fmt.Printf("Memory (%s): total=%d free=%d\n",
        r.Hostname, r.Memory.Total, r.Memory.Free)
}
```

## Example

Refer to
[`examples/sdk/client/memory.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/memory.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
