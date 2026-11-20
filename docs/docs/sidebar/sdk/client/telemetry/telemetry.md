---
sidebar_position: 5
---

# Telemetry

Telemetry that the controller reads for you, such as the spans of a trace.

## Methods

| Method                | Description                                                                           |
| --------------------- | ------------------------------------------------------------------------------------- |
| `Trace(ctx, id, all)` | Retrieve a trace by its ID: its summary and its span tree. `all` keeps the NATS calls |

## Usage

```go
// The trace ID that a command prints after its job ID, or resp.TraceID()
resp, err := client.Telemetry.Trace(ctx, "26e448c9dbe1c4a8d216a98c675079c9", false)
fmt.Println(resp.Data.Duration, resp.Data.NATSCalls, "NATS calls left out")
for _, root := range resp.Data.Spans {
    // Each span has its children, in start order.
    fmt.Println(root.Name, root.On, root.Status, root.Detail, len(root.Children))
}
```

The controller reads the trace from the tracing backend at
`telemetry.tracing.url`. The backend gets the spans a few seconds after the
request. Thus a trace of a request that ran a moment ago can return a
`NotFoundError`. Try again after a few seconds. A controller with no backend
returns an error that names the setting.

## Example

Refer to
[`examples/sdk/client/trace.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/trace.go)
for a complete working example.

## Permissions

Requires `trace:read` permission.
