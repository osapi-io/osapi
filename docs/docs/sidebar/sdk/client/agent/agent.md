---
sidebar_position: 1
---

# Agent

Agent discovery, details, and lifecycle management.

## Methods

| Method                           | Returns                        | Permission    | Description                   |
| -------------------------------- | ------------------------------ | ------------- | ----------------------------- |
| `c.Agent.List(ctx)`              | `Response[ListAgentsResponse]` | `agent:read`  | List active agents            |
| `c.Agent.Status(ctx)`            | `Response[ComponentList]`      | `health:read` | Show the health of each agent |
| `c.Agent.Get(ctx, hostname)`     | `Response[AgentInfo]`          | `agent:read`  | Get agent details             |
| `c.Agent.Drain(ctx, hostname)`   | `error`                        | `agent:write` | Drain an agent                |
| `c.Agent.Undrain(ctx, hostname)` | `error`                        | `agent:write` | Undrain an agent              |
| `c.Facts.List(ctx)`              | `Response[FactKeysResponse]`   | `agent:read`  | List available fact keys      |

## Types

The fields of each type are the properties of its schema in the API's spec.

### ListAgentsResponse

| Field        | Type          | Description                    |
| ------------ | ------------- | ------------------------------ |
| `Items`      | `[]AgentInfo` | The active agents.             |
| `TotalItems` | `int`         | Total number of active agents. |

### ComponentList

| Field        | Type               | Description                                          |
| ------------ | ------------------ | ---------------------------------------------------- |
| `Items`      | `[]ComponentEntry` | Each component of the type, with its sub-components. |
| `TotalItems` | `int`              | The number of components of the type.                |

### AgentInfo

| Field              | Type                         | Description                                                                                                                                                                                                                   |
| ------------------ | ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Architecture`     | `string`                     | CPU architecture.                                                                                                                                                                                                             |
| `Conditions`       | `[]NodeCondition`            | The conditions that hold now, such as HighLoad. A condition that does not hold is not in the list.                                                                                                                            |
| `CPUCount`         | `int`                        | Number of logical CPUs.                                                                                                                                                                                                       |
| `Facts`            | `map[string]any`             | Extended facts from additional providers.                                                                                                                                                                                     |
| `Fingerprint`      | `string`                     | The SHA256 fingerprint of the agent's PKI public key. Empty when PKI is disabled.                                                                                                                                             |
| `FQDN`             | `string`                     | Fully qualified domain name.                                                                                                                                                                                                  |
| `Hostname`         | `string`                     | The hostname of the agent.                                                                                                                                                                                                    |
| `Interfaces`       | `[]NetworkInterfaceResponse` | Interfaces is the property interfaces.                                                                                                                                                                                        |
| `KernelVersion`    | `string`                     | OS kernel version.                                                                                                                                                                                                            |
| `KeyStored`        | `bool`                       | Whether the controller holds this agent's public key, recorded when its enrollment was accepted. False for an agent that has not enrolled since the key store existed, which is what to re-enrol before enabling enforcement. |
| `Labels`           | `map[string]string`          | Key-value labels configured on the agent.                                                                                                                                                                                     |
| `LoadAverage`      | `LoadAverageResponse`        | The system load averages for 1, 5, and 15 minutes.                                                                                                                                                                            |
| `MachineID`        | `string`                     | The permanent machine identifier.                                                                                                                                                                                             |
| `Memory`           | `MemoryResponse`             | Memory usage information.                                                                                                                                                                                                     |
| `MissingPrograms`  | `[]string`                   | The programs that the agent's behaviors declare and that the host lacks. A behavior fails until its program is installed.                                                                                                     |
| `OSInfo`           | `OSInfoResponse`             | Operating system information.                                                                                                                                                                                                 |
| `PackageMgr`       | `string`                     | Package manager.                                                                                                                                                                                                              |
| `PrimaryInterface` | `string`                     | Name of the interface used for the default route.                                                                                                                                                                             |
| `RegisteredAt`     | `time.Time`                  | When the agent last refreshed its heartbeat.                                                                                                                                                                                  |
| `Routes`           | `[]RouteResponse`            | Network routing table entries.                                                                                                                                                                                                |
| `ServiceMgr`       | `string`                     | Init system.                                                                                                                                                                                                                  |
| `StartedAt`        | `time.Time`                  | When the agent process started.                                                                                                                                                                                               |
| `State`            | `string`                     | The scheduling state of the agent: Ready, Draining, Cordoned or Pending. A drain that the agent has not applied yet is Draining. It is one of: Ready, Draining, Cordoned, Pending.                                            |
| `Timeline`         | `[]TimelineEvent`            | Agent state transition history.                                                                                                                                                                                               |
| `Uptime`           | `string`                     | The system uptime.                                                                                                                                                                                                            |
| `Verified`         | `bool`                       | Whether this agent's registration was signed by its stored key and claims the hostname it enrolled under. Only a verified registration decides where work is routed. True when the controller is not enforcing verification.  |
| `Version`          | `string`                     | The build version of the agent binary. Absent for an agent that does not report it.                                                                                                                                           |

### FactKeysResponse

| Field        | Type             | Description                                |
| ------------ | ---------------- | ------------------------------------------ |
| `Items`      | `[]FactKeyEntry` | Available fact keys for @fact. references. |
| `TotalItems` | `int`            | Total number of fact keys.                 |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all agents
resp, err := c.Agent.List(ctx)
for _, a := range resp.Data.Items {
    fmt.Printf("%s  state=%s\n", a.Hostname, a.State)
}

// Get specific agent details
resp, err := c.Agent.Get(ctx, "web-01")

// Drain an agent (stop new jobs)
err := c.Agent.Drain(ctx, "web-01")

// Undrain an agent (resume jobs)
err := c.Agent.Undrain(ctx, "web-01")
```

## Examples

Refer to
[`examples/sdk/client/agent.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/agent.go)
for agent discovery and facts.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method. Drain and Undrain return only an error: the API answers 204. To admit,
reject or revoke a machine, use the [Admission](admission.md) service.
