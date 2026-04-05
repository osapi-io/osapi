# Network Interface & Route Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add full CRUD for network interface configuration and route management
via Netplan drop-in files, following the direct-write provider pattern
established by sysctl and DNS.

**Architecture:** A new `netplan` provider in
`internal/provider/network/netplan/` handles interface and route CRUD. It
generates Netplan YAML, writes drop-in files with `osapi-` prefix, validates
with `netplan generate`, and applies with `netplan apply`. List/get for
interfaces reuses the existing `netinfo` provider for system state. The shared
`netplan.ApplyConfig` helper (from DNS migration) handles the write → validate →
apply flow.

**Tech Stack:** Go, Netplan YAML, NATS JetStream KV, avfs

**Baseline coverage:** 99.9% — must not regress.

---

### Task 1: Interface and route provider types

**Files:**

- Create: `internal/provider/network/netplan/types.go`
- Create: `internal/provider/network/netplan/mocks/generate.go`

- [ ] **Step 1: Define the provider interfaces and types**

Create `internal/provider/network/netplan/types.go`:

```go
package netplan

import "context"

// InterfaceProvider manages network interface configuration via Netplan.
type InterfaceProvider interface {
    List(ctx context.Context) ([]InterfaceEntry, error)
    Get(ctx context.Context, name string) (*InterfaceEntry, error)
    Create(ctx context.Context, entry InterfaceEntry) (*InterfaceResult, error)
    Update(ctx context.Context, entry InterfaceEntry) (*InterfaceResult, error)
    Delete(ctx context.Context, name string) (*InterfaceResult, error)
}

// RouteProvider manages route configuration via Netplan.
type RouteProvider interface {
    List(ctx context.Context) ([]RouteListEntry, error)
    Get(ctx context.Context, interfaceName string) (*RouteEntry, error)
    Create(ctx context.Context, entry RouteEntry) (*RouteResult, error)
    Update(ctx context.Context, entry RouteEntry) (*RouteResult, error)
    Delete(ctx context.Context, interfaceName string) (*RouteResult, error)
}

// InterfaceEntry represents a managed interface configuration.
type InterfaceEntry struct {
    Name       string   `json:"name"`
    DHCP4      *bool    `json:"dhcp4,omitempty"`
    DHCP6      *bool    `json:"dhcp6,omitempty"`
    Addresses  []string `json:"addresses,omitempty"`
    Gateway4   string   `json:"gateway4,omitempty"`
    Gateway6   string   `json:"gateway6,omitempty"`
    MTU        int      `json:"mtu,omitempty"`
    MACAddress string   `json:"mac_address,omitempty"`
    WakeOnLAN  *bool    `json:"wakeonlan,omitempty"`
    Managed    bool     `json:"managed,omitempty"`
}

// InterfaceResult is the outcome of a create/update/delete.
type InterfaceResult struct {
    Name    string `json:"name"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}

// RouteEntry represents managed routes for an interface.
type RouteEntry struct {
    Interface string  `json:"interface"`
    Routes    []Route `json:"routes"`
}

// Route is a single route definition.
type Route struct {
    To     string `json:"to"`
    Via    string `json:"via"`
    Metric int    `json:"metric,omitempty"`
}

// RouteListEntry is a route from the system routing table.
type RouteListEntry struct {
    Destination string `json:"destination"`
    Gateway     string `json:"gateway"`
    Interface   string `json:"interface"`
    Mask        string `json:"mask,omitempty"`
    Metric      int    `json:"metric,omitempty"`
    Flags       string `json:"flags,omitempty"`
}

// RouteResult is the outcome of a route create/update/delete.
type RouteResult struct {
    Interface string `json:"interface"`
    Changed   bool   `json:"changed"`
    Error     string `json:"error,omitempty"`
}
```

- [ ] **Step 2: Add mock generation**

Create `internal/provider/network/netplan/mocks/generate.go`:

```go
package mocks

//go:generate go tool github.com/golang/mock/mockgen -source=../types.go -destination=provider.gen.go -package=mocks
```

Run: `go generate ./internal/provider/network/netplan/mocks/...`

- [ ] **Step 3: Verify build**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```
feat(netplan): add interface and route provider types
```

---

### Task 2: Interface provider implementation

**Files:**

- Create: `internal/provider/network/netplan/interface.go`
- Create: `internal/provider/network/netplan/interface_public_test.go`
- Create: `internal/provider/network/netplan/debian.go`
- Create: `internal/provider/network/netplan/darwin.go`
- Create: `internal/provider/network/netplan/linux.go`

- [ ] **Step 1: Create the Debian interface provider**

Create `internal/provider/network/netplan/debian.go` with the struct and
constructor:

```go
type Debian struct {
    provider.FactsAware
    logger      *slog.Logger
    fs          avfs.VFS
    stateKV     jetstream.KeyValue
    execManager exec.Manager
    hostname    string
    netinfo     netinfo.Provider
}

func NewDebianProvider(
    logger *slog.Logger,
    fs avfs.VFS,
    stateKV jetstream.KeyValue,
    execManager exec.Manager,
    hostname string,
    netinfo netinfo.Provider,
) *Debian
```

Compile-time checks for `InterfaceProvider`, `RouteProvider`, and
`provider.FactsSetter`.

Create `darwin.go` and `linux.go` stubs returning `ErrUnsupported`.

- [ ] **Step 2: Implement interface CRUD**

Create `internal/provider/network/netplan/interface.go`:

**List** — delegates to `netinfo.GetInterfaces()`. For each interface, check if
an `osapi-{name}.yaml` file exists to set `Managed: true`.

**Get** — delegates to `netinfo.GetInterfaces()`, filters by name. Checks
managed status.

**Create** — validates name, checks file doesn't exist, generates Netplan YAML,
calls `ApplyConfig`.

**Update** — validates name, checks file exists, generates Netplan YAML, calls
`ApplyConfig`.

**Delete** — calls `RemoveConfig`.

YAML generation function `generateInterfaceYAML(entry InterfaceEntry)`:

```yaml
network:
  version: 2
  ethernets:
    eth0:
      dhcp4: false
      dhcp6: false
      addresses:
        - 10.0.0.5/24
      gateway4: 10.0.0.1
      mtu: 1500
```

File path: `/etc/netplan/osapi-{name}.yaml`

- [ ] **Step 3: Write interface tests**

Create `internal/provider/network/netplan/interface_public_test.go`. Use
testify/suite, table-driven, validateFunc. Use `memfs`, gomock.

Test each method: List (with managed + unmanaged), Get (found, not found),
Create (success, already exists, generate fails), Update (success, not found),
Delete (success, not found). YAML generation tests for each field combination.

Target: 100% coverage on `interface.go`.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/provider/network/netplan/... -count=1`

- [ ] **Step 5: Commit**

```
feat(netplan): implement interface CRUD with Netplan
```

---

### Task 3: Route provider implementation

**Files:**

- Create: `internal/provider/network/netplan/route.go`
- Create: `internal/provider/network/netplan/route_public_test.go`

- [ ] **Step 1: Implement route CRUD**

Create `internal/provider/network/netplan/route.go`:

**List** — delegates to `netinfo.GetRoutes()`. Converts `RouteResult` to
`RouteListEntry`.

**Get** — reads the managed route file from state KV or disk for the given
interface. Parses the YAML to extract routes.

**Create** — validates interface name, checks file doesn't exist, validates no
default route in list, generates YAML, calls `ApplyConfig`.

**Update** — same as create but file must exist.

**Delete** — validates no default route in managed routes, calls `RemoveConfig`.

YAML generation function `generateRouteYAML(entry RouteEntry)`:

```yaml
network:
  version: 2
  ethernets:
    eth0:
      routes:
        - to: 10.1.0.0/16
          via: 10.0.0.1
          metric: 100
```

File path: `/etc/netplan/osapi-{interface}-routes.yaml`

Default route protection: reject create/update if any route has `To` of
`0.0.0.0/0`, `::/0`, or `default`.

- [ ] **Step 2: Write route tests**

Create `internal/provider/network/netplan/route_public_test.go`.

Test each method. Include default route protection tests (reject `0.0.0.0/0`).
YAML generation tests.

Target: 100% coverage on `route.go`.

- [ ] **Step 3: Run tests**

Run: `go test ./internal/provider/network/netplan/... -count=1`

- [ ] **Step 4: Commit**

```
feat(netplan): implement route CRUD with default route protection
```

---

### Task 4: Job operations and agent processor

**Files:**

- Modify: `pkg/sdk/client/operations.go`
- Modify: `internal/job/types.go`
- Create: `internal/agent/processor_interface.go`
- Create: `internal/agent/processor_route.go`
- Create: `internal/agent/processor_interface_public_test.go`
- Create: `internal/agent/processor_route_public_test.go`
- Modify: `internal/agent/processor_network.go`

- [ ] **Step 1: Add operation constants**

In `pkg/sdk/client/operations.go`, add:

```go
// Network interface operations.
const (
    OpNetworkInterfaceList   JobOperation = "interface.list"
    OpNetworkInterfaceGet    JobOperation = "interface.get"
    OpNetworkInterfaceCreate JobOperation = "interface.create"
    OpNetworkInterfaceUpdate JobOperation = "interface.update"
    OpNetworkInterfaceDelete JobOperation = "interface.delete"
)

// Network route operations.
const (
    OpNetworkRouteList   JobOperation = "route.list"
    OpNetworkRouteGet    JobOperation = "route.get"
    OpNetworkRouteCreate JobOperation = "route.create"
    OpNetworkRouteUpdate JobOperation = "route.update"
    OpNetworkRouteDelete JobOperation = "route.delete"
)
```

Mirror in `internal/job/types.go`.

- [ ] **Step 2: Create interface processor**

Create `internal/agent/processor_interface.go` with `processInterfaceOperation`
that dispatches list/get/create/update/delete to the provider. Follow existing
processor patterns (e.g., `processor_sysctl.go`).

- [ ] **Step 3: Create route processor**

Create `internal/agent/processor_route.go` with `processRouteOperation`. Same
pattern.

- [ ] **Step 4: Wire into network processor**

In `internal/agent/processor_network.go`, update `NewNetworkProcessor` to accept
`InterfaceProvider` and `RouteProvider`. Add `case "interface"` and
`case "route"` to the switch.

- [ ] **Step 5: Write processor tests**

Create test files for both processors. Follow existing patterns.

- [ ] **Step 6: Run tests**

Run: `go test ./internal/agent/... -count=1` Run: `go build ./...`

- [ ] **Step 7: Commit**

```
feat(network): add interface and route agent processors
```

---

### Task 5: Agent wiring

**Files:**

- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Create and register providers**

In `cmd/agent_setup.go`, create the Netplan provider (Debian only,
ErrUnsupported on other platforms) and pass it to `NewNetworkProcessor`
alongside the existing DNS and ping providers.

The Netplan provider needs `fs`, `stateKV`, `execManager`, `hostname`, and
`netinfoProvider` — all already available in agent setup.

- [ ] **Step 2: Verify build**

Run: `go build ./...`

- [ ] **Step 3: Commit**

```
feat(network): wire interface and route providers in agent
```

---

### Task 6: OpenAPI spec and code generation

**Files:**

- Modify: `internal/controller/api/node/network/gen/api.yaml`

- [ ] **Step 1: Add interface endpoints to the OpenAPI spec**

Add to the existing network spec:

```
GET    /node/{hostname}/network/interface
GET    /node/{hostname}/network/interface/{name}
POST   /node/{hostname}/network/interface/{name}
PUT    /node/{hostname}/network/interface/{name}
DELETE /node/{hostname}/network/interface/{name}
```

Request body for POST/PUT (`InterfaceConfigRequest`):

- `dhcp4` (bool, omitempty)
- `dhcp6` (bool, omitempty)
- `addresses` ([]string, omitempty, dive, cidr)
- `gateway4` (string, omitempty, ipv4)
- `gateway6` (string, omitempty, ipv6)
- `mtu` (int, omitempty, min=68, max=9000)
- `mac_address` (string, omitempty)
- `wakeonlan` (bool, omitempty)

All with `x-oapi-codegen-extra-tags` validate tags.

- [ ] **Step 2: Add route endpoints**

```
GET    /node/{hostname}/network/route
GET    /node/{hostname}/network/route/{interface}
POST   /node/{hostname}/network/route/{interface}
PUT    /node/{hostname}/network/route/{interface}
DELETE /node/{hostname}/network/route/{interface}
```

Request body for POST/PUT (`RouteConfigRequest`):

- `routes` ([]RouteItem, required, min=1)

  - `to` (string, required, cidr)
  - `via` (string, required, ip)
  - `metric` (int, omitempty, min=0)

- [ ] **Step 3: Add DELETE for DNS**

Add `DELETE /node/{hostname}/network/dns` endpoint to remove managed DNS config.

- [ ] **Step 4: Regenerate code**

Run: `just generate`

- [ ] **Step 5: Commit**

```
feat(network): add interface, route, and DNS delete to OpenAPI spec
```

---

### Task 7: API handlers

**Files:**

- Create: `internal/controller/api/node/network/interface_list_get.go`
- Create: `internal/controller/api/node/network/interface_get.go`
- Create: `internal/controller/api/node/network/interface_create_post.go`
- Create: `internal/controller/api/node/network/interface_update_put.go`
- Create: `internal/controller/api/node/network/interface_delete.go`
- Create: `internal/controller/api/node/network/route_list_get.go`
- Create: `internal/controller/api/node/network/route_get.go`
- Create: `internal/controller/api/node/network/route_create_post.go`
- Create: `internal/controller/api/node/network/route_update_put.go`
- Create: `internal/controller/api/node/network/route_delete.go`
- Create: `internal/controller/api/node/network/dns_delete.go`
- Create: corresponding `*_public_test.go` for each handler

- [ ] **Step 1: Implement interface handlers**

Follow existing handler patterns (e.g., `sysctl` domain). Each handler:

- Validates hostname
- Validates request body (for POST/PUT)
- Calls `JobClient.Query`/`Modify` or broadcast variants
- Returns collection response

- [ ] **Step 2: Implement route handlers**

Same pattern. Route create/update validates request body.

- [ ] **Step 3: Implement DNS delete handler**

Calls `JobClient.Modify` with the delete operation.

- [ ] **Step 4: Write handler tests**

Unit tests + HTTP wiring tests + RBAC tests for each endpoint. Follow existing
test patterns in the network package.

- [ ] **Step 5: Update handler.go**

Update `Handler()` function — it should already pick up new endpoints from the
regenerated `StrictServerInterface`. Verify the compile-time check passes.

- [ ] **Step 6: Run tests**

Run: `go test ./internal/controller/api/node/network/... -count=1` Run:
`go build ./...`

- [ ] **Step 7: Commit**

```
feat(network): add interface, route, and DNS delete handlers
```

---

### Task 8: SDK service

**Files:**

- Create: `pkg/sdk/client/interface.go`
- Create: `pkg/sdk/client/interface_types.go`
- Create: `pkg/sdk/client/interface_public_test.go`
- Create: `pkg/sdk/client/interface_types_public_test.go`
- Create: `pkg/sdk/client/route.go`
- Create: `pkg/sdk/client/route_types.go`
- Create: `pkg/sdk/client/route_public_test.go`
- Create: `pkg/sdk/client/route_types_public_test.go`
- Modify: `pkg/sdk/client/dns.go` (add Delete method)
- Modify: `pkg/sdk/client/osapi.go` (add Interface and Route services)

- [ ] **Step 1: Create interface SDK service**

`InterfaceService` with List, Get, Create, Update, Delete methods. Follow
existing SDK patterns (e.g., `SysctlService`).

- [ ] **Step 2: Create route SDK service**

`RouteService` with List, Get, Create, Update, Delete methods.

- [ ] **Step 3: Add DNS Delete to SDK**

Add `Delete(ctx, target)` method to `DNSService`.

- [ ] **Step 4: Wire into Client**

Add `Interface *InterfaceService` and `Route *RouteService` fields to the
`Client` struct in `osapi.go`.

- [ ] **Step 5: Regenerate SDK client**

Run: `go generate ./pkg/sdk/client/gen/...`

- [ ] **Step 6: Write tests**

100% coverage on all SDK service methods and type conversions.

- [ ] **Step 7: Commit**

```
feat(sdk): add Interface and Route services
```

---

### Task 9: CLI commands

**Files:**

- Create: `cmd/client_node_network_interface.go`
- Create: `cmd/client_node_network_interface_list.go`
- Create: `cmd/client_node_network_interface_get.go`
- Create: `cmd/client_node_network_interface_create.go`
- Create: `cmd/client_node_network_interface_update.go`
- Create: `cmd/client_node_network_interface_delete.go`
- Create: `cmd/client_node_network_route.go`
- Create: `cmd/client_node_network_route_list.go`
- Create: `cmd/client_node_network_route_get.go`
- Create: `cmd/client_node_network_route_create.go`
- Create: `cmd/client_node_network_route_update.go`
- Create: `cmd/client_node_network_route_delete.go`
- Create: `cmd/client_node_network_dns_delete.go`

- [ ] **Step 1: Create interface CLI commands**

Parent `interface` command under `client node network`. Subcommands: list, get,
create, update, delete. Follow existing CLI patterns (flags for each field,
`--json` support, `printKV`/`printStyledTable`).

- [ ] **Step 2: Create route CLI commands**

Parent `route` command. Subcommands: list, get, create, update, delete.
`--route` flag accepts `to:via:metric` format for each route.

- [ ] **Step 3: Create DNS delete command**

`osapi client node network dns delete --target HOST`

- [ ] **Step 4: Verify build**

Run: `go build ./...`

- [ ] **Step 5: Commit**

```
feat(cli): add interface, route, and DNS delete commands
```

---

### Task 10: Documentation

**Files:**

- Create: `docs/docs/sidebar/features/network-interface-management.md`
- Create: `docs/docs/sidebar/usage/cli/client/node/network/interface/`
- Create: `docs/docs/sidebar/usage/cli/client/node/network/route/`
- Modify: `docs/docs/sidebar/features/network-management.md`
- Modify: `docs/docs/sidebar/features/features.md`
- Modify: `docs/docs/sidebar/architecture/architecture.md`
- Modify: `docs/docs/sidebar/architecture/api-guidelines.md`
- Modify: `docs/docusaurus.config.ts`
- Create: SDK doc pages and examples

- [ ] **Step 1: Create feature page**

Network interface and route management feature page with:

- Overview, how it works, Netplan drop-in pattern
- CLI examples for each operation
- Safety rules (default route protection)
- Managed file reference

- [ ] **Step 2: Create CLI doc pages**

One page per CLI subcommand with example output.

- [ ] **Step 3: Update cross-references**

Features page, architecture, API guidelines, navbar dropdown.

- [ ] **Step 4: Create SDK docs and examples**

SDK doc page for Interface and Route services. Example files.

- [ ] **Step 5: Commit**

```
docs: add network interface and route management documentation
```

---

### Task 11: Final verification

- [ ] **Step 1: Run full test suite**

```bash
just go::unit
```

- [ ] **Step 2: Build and lint**

```bash
go build ./...
just go::vet
```

- [ ] **Step 3: Coverage check**

```bash
just go::unit-cov 2>&1 | tail -1
```

Must be >= 99.9%.

- [ ] **Step 4: Cross-layer consistency check**

Verify the interface and route domains appear in all the same places as existing
domains (grep for "sysctl" across the codebase and confirm "interface" and
"route" appear in the same files).
