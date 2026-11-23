---
sidebar_position: 3
---

# Health

Health check operations.

## Methods

| Method                     | Returns                    | Permission    | Description                        |
| -------------------------- | -------------------------- | ------------- | ---------------------------------- |
| `c.Health.Liveness(ctx)`   | `Response[HealthResponse]` | none          | Liveness probe                     |
| `c.Health.Ready(ctx)`      | `Response[ReadyResponse]`  | none          | Readiness probe                    |
| `c.Health.Status(ctx)`     | `Response[StatusResponse]` | `health:read` | System status and component health |
| `c.Controller.Status(ctx)` | `Response[ComponentList]`  | `health:read` | Show the health of the controller  |
| `c.NATS.Status(ctx)`       | `Response[ComponentList]`  | `health:read` | Show the health of the NATS server |

## Types

The fields of each type are the properties of its schema in the API's spec.

### HealthResponse

| Field    | Type     | Description    |
| -------- | -------- | -------------- |
| `Status` | `string` | Health status. |

### ReadyResponse

| Field    | Type     | Description                   |
| -------- | -------- | ----------------------------- |
| `Error`  | `string` | Error message when not ready. |
| `Status` | `string` | Readiness status.             |

### StatusResponse

| Field          | Type                | Description                                                                                            |
| -------------- | ------------------- | ------------------------------------------------------------------------------------------------------ |
| `Agents`       | `AgentStats`        | Agents is the property agents.                                                                         |
| `Checks`       | `[]Check`           | What the controller that answered checked for this answer: its connection to NATS and to the KV store. |
| `Consumers`    | `int`               | Total number of JetStream consumers: the sum of the consumers of each stream.                          |
| `Jobs`         | `JobStats`          | Jobs is the property jobs.                                                                             |
| `KVBuckets`    | `[]KVBucketInfo`    | KV bucket statistics.                                                                                  |
| `NATS`         | `NATSInfo`          | NATS is the property nats.                                                                             |
| `ObjectStores` | `[]ObjectStoreInfo` | Object Store statistics.                                                                               |
| `Registry`     | `[]ComponentEntry`  | All registered components with health details, each with its sub-components.                           |
| `Status`       | `string`            | Overall health status.                                                                                 |
| `Streams`      | `[]StreamInfo`      | JetStream stream statistics.                                                                           |
| `Uptime`       | `string`            | Time since server started.                                                                             |
| `Version`      | `string`            | Application version.                                                                                   |

### ComponentList

| Field        | Type               | Description                                          |
| ------------ | ------------------ | ---------------------------------------------------- |
| `Items`      | `[]ComponentEntry` | Each component of the type, with its sub-components. |
| `TotalItems` | `int`              | The number of components of the type.                |

## Usage

```go
// Simple liveness check (unauthenticated)
resp, err := c.Health.Liveness(ctx)

// Readiness check (unauthenticated)
resp, err := c.Health.Ready(ctx)

// Detailed status (requires auth)
resp, err := c.Health.Status(ctx)
```

## Example

Refer to
[`examples/sdk/client/health.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/health.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method. A method with none needs no token.
