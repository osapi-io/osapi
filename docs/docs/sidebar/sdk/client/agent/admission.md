---
sidebar_position: 2
---

# Admission

Admission decides which machines can join. When PKI is on, a new agent asks to
join, and its request waits until an operator admits or rejects it. An admitted
machine can be revoked later.

## Methods

| Method                                                | Returns                           | Permission        | Description                      |
| ----------------------------------------------------- | --------------------------------- | ----------------- | -------------------------------- |
| `c.Admission.List(ctx, params AdmissionListParams)`   | `Response[ListAdmissionResponse]` | `admission:read`  | List admission records           |
| `c.Admission.Get(ctx, machineID)`                     | `Response[Admission]`             | `admission:read`  | Get a machine's admission record |
| `c.Admission.Delete(ctx, machineID)`                  | `error`                           | `admission:write` | Delete a machine                 |
| `c.Admission.Admit(ctx, input AdmissionAdmitInput)`   | `error`                           | `admission:write` | Admit a pending machine          |
| `c.Admission.Reject(ctx, input AdmissionRejectInput)` | `error`                           | `admission:write` | Reject a pending machine         |
| `c.Admission.Revoke(ctx, machineID)`                  | `error`                           | `admission:write` | Revoke an admitted machine       |

## Types

The fields of each type are the properties of its schema in the API's spec.

### AdmissionListParams

| Field   | Type     | Description                                                             |
| ------- | -------- | ----------------------------------------------------------------------- |
| `State` | `string` | Only the records in this state: pending, admitted, rejected or revoked. |

### ListAdmissionResponse

| Field        | Type          | Description                               |
| ------------ | ------------- | ----------------------------------------- |
| `Items`      | `[]Admission` | The admission records, in hostname order. |
| `TotalItems` | `int`         | The number of records.                    |

### Admission

| Field                 | Type        | Description                                                                               |
| --------------------- | ----------- | ----------------------------------------------------------------------------------------- |
| `AdmittedAt`          | `time.Time` | When the node was admitted.                                                               |
| `DecidedAt`           | `time.Time` | When the node was rejected or revoked.                                                    |
| `Fingerprint`         | `string`    | The fingerprint of the node's key.                                                        |
| `Hostname`            | `string`    | The hostname the node gave.                                                               |
| `LastSeen`            | `time.Time` | When the node last reported.                                                              |
| `MachineID`           | `string`    | The machine ID of the node.                                                               |
| `Online`              | `bool`      | Whether the node is running now: it is in the registry.                                   |
| `Previous`            | `string`    | On a pending record, the state of the node's earlier record: a returning node.            |
| `PreviousFingerprint` | `string`    | On a pending record, the fingerprint of the earlier record; a different one is a new key. |
| `Reason`              | `string`    | Why the node was rejected.                                                                |
| `RequestedAt`         | `time.Time` | When the node asked to join.                                                              |
| `State`               | `string`    | Pending, admitted, rejected or revoked.                                                   |

### AdmissionAdmitInput

| Field         | Type     | Description                                                           |
| ------------- | -------- | --------------------------------------------------------------------- |
| `MachineID`   | `string` | The machine_id of the path: the machine ID of the node.               |
| `Fingerprint` | `string` | The fingerprint that osapi agent admission get prints on the machine. |

### AdmissionRejectInput

| Field       | Type     | Description                                                      |
| ----------- | -------- | ---------------------------------------------------------------- |
| `MachineID` | `string` | The machine_id of the path: the machine ID of the node.          |
| `Reason`    | `string` | Why the machine is rejected; the machine and the record keep it. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List the machines that wait for a decision
resp, err := c.Admission.List(ctx, client.AdmissionListParams{State: "pending"})
for _, a := range resp.Data.Items {
    fmt.Printf("%s  %s  %s\n", a.MachineID, a.Hostname, a.Fingerprint)
}

// Admit, with the fingerprint that `osapi agent admission get` prints on the
// machine
err := c.Admission.Admit(ctx, client.AdmissionAdmitInput{
    MachineID: "d530664edc1c431eb5ef1e9dfd6617a5",
    Fingerprint: "SHA256:1f53dea6...",
})

// Reject, with the reason
err := c.Admission.Reject(ctx, client.AdmissionRejectInput{
    MachineID: "d530664edc1c431eb5ef1e9dfd6617a5",
    Reason: "unknown host",
})

// Revoke a machine that is not trusted: nothing that it signs verifies
// after this, and a record shows it if it asks again
err := c.Admission.Revoke(ctx, "d530664edc1c431eb5ef1e9dfd6617a5")

// Delete a machine that goes away on purpose, such as before a rebuild: it
// leaves no record, and is new if it asks again
err := c.Admission.Delete(ctx, "d530664edc1c431eb5ef1e9dfd6617a5")
```

Admit compares the fingerprint with the key in the request of the machine. A
different fingerprint returns a `ConflictError`, and the machine stays pending.
Read the fingerprint on the machine, not from `List`: the purpose of the check
is to show that the request came from that machine.

## Example

Refer to
[`examples/sdk/client/admission.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/admission.go).

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method. Admit, Reject, Revoke and Delete return only an error: the API
answers 204. When PKI is off, each method returns the 503 error of the API.
