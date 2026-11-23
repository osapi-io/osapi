---
sidebar_position: 5
---

# Load

Load average query operations.

## Methods

| Method                      | Returns                           | Permission  | Description                       |
| --------------------------- | --------------------------------- | ----------- | --------------------------------- |
| `c.Load.Get(ctx, hostname)` | `Response[Collection[LoadEntry]]` | `load:read` | Get the load averages of the host |

## Types

The fields of each type are the properties of its schema in the API's spec.

### LoadEntry

| Field         | Type       | Description                                                                            |
| ------------- | ---------- | -------------------------------------------------------------------------------------- |
| `Error`       | `string`   | Error message if the agent failed.                                                     |
| `Hostname`    | `string`   | Hostname of the agent that produced this result.                                       |
| `LoadAverage` | `Averages` | LoadAverage is the property load_average.                                              |
| `Status`      | `string`   | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get load averages from all hosts
resp, err := c.Load.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    fmt.Printf("Load (%s): %.2f %.2f %.2f\n",
        r.Hostname,
        r.LoadAverage.Min1,
        r.LoadAverage.Min5,
        r.LoadAverage.Min15)
}
```

## Example

Refer to
[`examples/sdk/client/load.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/load.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
