---
sidebar_position: 7
---

# MetricsService

Prometheus metrics access.

## Methods

| Method     | Description                       |
| ---------- | --------------------------------- |
| `Get(ctx)` | Fetch raw Prometheus metrics text |

## Usage

```go
text, err := client.Metrics.Get(ctx)
fmt.Print(text)
```

## Example

See
[`examples/sdk/osapi/metrics.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/osapi/metrics.go)
for a complete working example.

## Permissions

Unauthenticated. The `/metrics` endpoint is open.
