---
sidebar_position: 5
---

# Telemetry

Telemetry that the controller reads for you, such as the spans of a trace.

## Methods

| Method                                                | Returns                   | Permission   | Description              |
| ----------------------------------------------------- | ------------------------- | ------------ | ------------------------ |
| `c.Telemetry.Get(ctx, id, params TelemetryGetParams)` | `Response[TraceResponse]` | `trace:read` | Get the spans of a trace |

## Types

The fields of each type are the properties of its schema in the API's spec.

### TelemetryGetParams

| Field | Type   | Description                                                     |
| ----- | ------ | --------------------------------------------------------------- |
| `All` | `bool` | Show the NATS calls too, the reads and writes of the job queue. |

### TraceResponse

| Field         | Type          | Description                                                                          |
| ------------- | ------------- | ------------------------------------------------------------------------------------ |
| `Duration`    | `string`      | From the start of the first span to the end of the last, as a Go duration.           |
| `FailedSpans` | `int`         | The number of spans that failed.                                                     |
| `NATSCalls`   | `int`         | The number of NATS calls left out, the bookkeeping of the job queue. all shows them. |
| `NATSTime`    | `string`      | The time that the NATS calls left out took, as a Go duration.                        |
| `Services`    | `[]string`    | The services that made the spans, in the order each first appears.                   |
| `Spans`       | `[]TraceNode` | The spans that have no parent in the trace, each with its children, in start order.  |
| `Started`     | `time.Time`   | When the first span started.                                                         |
| `TraceID`     | `string`      | The trace ID.                                                                        |

## Usage

```go
// The trace ID that a command prints after its job ID, or resp.TraceID()
resp, err := c.Telemetry.Get(ctx, "26e448c9dbe1c4a8d216a98c675079c9",
    client.TelemetryGetParams{})
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

The Permission column of [Methods](#methods) gives the permission of each
method.
