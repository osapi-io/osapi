---
sidebar_position: 4
---

# Audit

Audit log operations.

## Methods

| Method                                      | Returns                        | Permission   | Description                  |
| ------------------------------------------- | ------------------------------ | ------------ | ---------------------------- |
| `c.Audit.List(ctx, params AuditListParams)` | `Response[ListAuditResponse]`  | `audit:read` | List audit log entries       |
| `c.Audit.Export(ctx)`                       | `Response[ListAuditResponse]`  | `audit:read` | Export all audit log entries |
| `c.Audit.Get(ctx, id)`                      | `Response[AuditEntryResponse]` | `audit:read` | Get a single audit log entry |

## Types

The fields of each type are the properties of its schema in the API's spec.

### AuditListParams

| Field       | Type     | Description                                                                  |
| ----------- | -------- | ---------------------------------------------------------------------------- |
| `Limit`     | `int`    | Maximum number of entries to return.                                         |
| `Offset`    | `int`    | Number of entries to skip.                                                   |
| `Writes`    | `bool`   | Only the requests that change something: each method but GET.                |
| `User`      | `string` | Only the entries of this user, the subject of the token.                     |
| `Operation` | `string` | Only the entries whose operation ID or path has this text, such as hostname. |
| `Status`    | `int`    | Only the entries with this response status, such as 403.                     |

### ListAuditResponse

| Field        | Type           | Description                      |
| ------------ | -------------- | -------------------------------- |
| `Items`      | `[]AuditEntry` | The audit entries for this page. |
| `TotalItems` | `int`          | Total number of audit entries.   |

### AuditEntryResponse

| Field   | Type         | Description                  |
| ------- | ------------ | ---------------------------- |
| `Entry` | `AuditEntry` | Entry is the property entry. |

## Usage

```go
// List recent entries
resp, err := c.Audit.List(ctx, client.AuditListParams{Limit: 20})

// Get a specific entry
resp, err := c.Audit.Get(ctx, "uuid-string")

// Export all entries
resp, err := c.Audit.Export(ctx)
```

## Example

Refer to
[`examples/sdk/client/audit.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/audit.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
