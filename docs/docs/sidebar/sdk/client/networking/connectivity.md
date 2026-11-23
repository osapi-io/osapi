---
sidebar_position: 2
---

# Connectivity

Whether a host reaches an address, by ICMP ping: the connectivity behavior.

## Methods

| Method                                                                          | Returns                                               | Permission             | Description                   |
| ------------------------------------------------------------------------------- | ----------------------------------------------------- | ---------------------- | ----------------------------- |
| `c.NetworkConnectivity.Ping(ctx, hostname, input NetworkConnectivityPingInput)` | `Response[Collection[NetworkConnectivityPingResult]]` | `connectivity:execute` | Ping an address from the host |

## Types

The fields of each type are the properties of its schema in the API's spec.

### NetworkConnectivityPingInput

| Field     | Type     | Description                                                                       |
| --------- | -------- | --------------------------------------------------------------------------------- |
| `Address` | `string` | The IPv4 or IPv6 address to ping, or an @fact. reference that the agent resolves. |

### NetworkConnectivityPingResult

| Field             | Type      | Description                                                                                 |
| ----------------- | --------- | ------------------------------------------------------------------------------------------- |
| `AvgRTT`          | `string`  | Average round-trip time in milliseconds, with two decimals. Absent when no reply came back. |
| `Changed`         | `bool`    | Always false: a ping changes nothing on the host.                                           |
| `Error`           | `string`  | Error message if the agent failed.                                                          |
| `Hostname`        | `string`  | Hostname of the agent that produced this result.                                            |
| `MaxRTT`          | `string`  | Maximum round-trip time in milliseconds, with two decimals. Absent when no reply came back. |
| `MinRTT`          | `string`  | Minimum round-trip time in milliseconds, with two decimals. Absent when no reply came back. |
| `PacketLoss`      | `float64` | Percentage of packet loss.                                                                  |
| `PacketsReceived` | `int`     | Number of packets received.                                                                 |
| `PacketsSent`     | `int`     | Number of packets sent.                                                                     |
| `Status`          | `string`  | The status of the operation for this host. It is one of: ok, failed, skipped, timeout.      |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Ping an address from all agents
resp, err := c.NetworkConnectivity.Ping(ctx, "_all",
    client.NetworkConnectivityPingInput{Address: "8.8.8.8"})
for _, r := range resp.Data.Results {
    fmt.Printf("Ping (%s): sent=%d received=%d loss=%.1f%%\n",
        r.Hostname, r.PacketsSent, r.PacketsReceived, r.PacketLoss)
}
```

A host that does not answer is 100% loss, not an error.

## Example

Refer to
[`examples/sdk/client/connectivity.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/connectivity.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
