---
sidebar_position: 2
---

# Admission

Admission decides which machines can join. When PKI is on, a new agent asks to
join, and its request waits until an operator admits or rejects it. An admitted
machine can be revoked later.

## Methods

| Method                               | Description                             |
| ------------------------------------ | --------------------------------------- |
| `List(ctx, state)`                   | List the records; `""` lists each state |
| `Get(ctx, machineID)`                | Get the record of one machine           |
| `Admit(ctx, machineID, fingerprint)` | Admit a pending machine                 |
| `Reject(ctx, machineID, reason)`     | Reject a pending machine                |
| `Revoke(ctx, machineID)`             | Remove the key of an admitted machine   |

The state is `pending`, `admitted`, `rejected` or `revoked`.

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List the machines that wait for a decision
resp, err := c.Admission.List(ctx, "pending")
for _, a := range resp.Data.Items {
    fmt.Printf("%s  %s  %s\n", a.MachineID, a.Hostname, a.Fingerprint)
}

// Admit, with the fingerprint that `osapi agent admission` prints on the
// machine
err := c.Admission.Admit(ctx, "d530664edc1c431eb5ef1e9dfd6617a5", "SHA256:1f53dea6...")

// Reject, with the reason
err := c.Admission.Reject(ctx, "d530664edc1c431eb5ef1e9dfd6617a5", "unknown host")

// Revoke: nothing that the machine signs verifies after this
err := c.Admission.Revoke(ctx, "d530664edc1c431eb5ef1e9dfd6617a5")
```

Admit compares the fingerprint with the key in the request of the machine. A
different fingerprint returns a `ConflictError`, and the machine stays pending.
Read the fingerprint on the machine, not from `List`: the purpose of the check
is to show that the request came from that machine.

## Example

Refer to
[`examples/sdk/client/admission.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/admission.go).

## Types

### `Admission`

| Field                 | Type         | Description                                      |
| --------------------- | ------------ | ------------------------------------------------ |
| `MachineID`           | `string`     | The machine ID, the key of the record            |
| `Hostname`            | `string`     | The hostname that the machine reported           |
| `Fingerprint`         | `string`     | The SHA256 fingerprint of the key of the machine |
| `State`               | `string`     | `pending`, `admitted`, `rejected` or `revoked`   |
| `Online`              | `bool`       | The agent reports now                            |
| `LastSeen`            | `*time.Time` | The last report of the agent                     |
| `RequestedAt`         | `*time.Time` | When the machine asked to join                   |
| `AdmittedAt`          | `*time.Time` | When the machine was admitted                    |
| `DecidedAt`           | `*time.Time` | When the machine was rejected or revoked         |
| `Reason`              | `string`     | Why the machine was rejected                     |
| `Previous`            | `string`     | The earlier state of a machine that asks again   |
| `PreviousFingerprint` | `string`     | The fingerprint of that earlier state            |

### `AdmissionList`

| Field        | Type          | Description           |
| ------------ | ------------- | --------------------- |
| `Items`      | `[]Admission` | The records           |
| `TotalItems` | `int`         | The number of records |

## Permissions

List and Get need `admission:read`. Admit, Reject and Revoke need
`admission:write`, and return only an error: the API answers 204. When PKI is
off, each method returns the 503 error of the API.
