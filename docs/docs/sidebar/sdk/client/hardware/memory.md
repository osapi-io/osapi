---
sidebar_position: 4
---

# Memory

Memory statistics query operations.

## Methods

| Method             | Description           |
| ------------------ | --------------------- |
| `Get(ctx, target)` | Get memory statistics |

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

See
[`examples/sdk/client/memory.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/memory.go)
for a complete working example.

## Permissions

Requires `node:read` permission.
