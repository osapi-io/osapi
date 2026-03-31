---
sidebar_position: 5
---

# Load

Load average query operations.

## Methods

| Method             | Description       |
| ------------------ | ----------------- |
| `Get(ctx, target)` | Get load averages |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get load averages from all hosts
resp, err := c.Load.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    fmt.Printf("Load (%s): %.2f %.2f %.2f\n",
        r.Hostname,
        r.LoadAverage.OneMin,
        r.LoadAverage.FiveMin,
        r.LoadAverage.FifteenMin)
}
```

## Example

See
[`examples/sdk/client/load.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/load.go)
for a complete working example.

## Permissions

Requires `node:read` permission.
