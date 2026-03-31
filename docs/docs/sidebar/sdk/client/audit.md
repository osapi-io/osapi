---
sidebar_position: 3
---

# Audit

Audit log operations.

## Methods

| Method                     | Description                      |
| -------------------------- | -------------------------------- |
| `List(ctx, limit, offset)` | Retrieve entries with pagination |
| `Get(ctx, id)`             | Retrieve a single entry by UUID  |
| `Export(ctx)`              | Retrieve all entries for export  |

## Usage

```go
// List recent entries
resp, err := client.Audit.List(ctx, 20, 0)

// Get a specific entry
resp, err := client.Audit.Get(ctx, "uuid-string")

// Export all entries
resp, err := client.Audit.Export(ctx)
```

## Example

See
[`examples/sdk/client/audit.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/audit.go)
for a complete working example.

## Permissions

Requires `audit:read` permission.
