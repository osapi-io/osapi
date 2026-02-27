---
title: Add MCP server for AI agent integration
status: backlog
created: 2026-02-18
updated: 2026-02-18
---

## Objective

Add an MCP (Model Context Protocol) server that exposes OSAPI's system and
network operations as MCP tools. This gives AI agents (Claude Code, Cursor,
etc.) native tool access to OSAPI without HTTP boilerplate.

The MCP server will be a fourth server-type command alongside `api server`,
`node agent`, and `nats server`, reusing the same `JobClient` layer and
`Lifecycle` pattern.

## Design

- **SDK**: `github.com/modelcontextprotocol/go-sdk` (official Go MCP SDK)
- **Transports**: stdio (default, for local AI tools) + streamable HTTP
  (flag-selected, for remote agents)
- **Tool scope**: System + Network operations only (no job queue internals)
- **CLI**: `osapi mcp server start [--transport http --port 3001]`

### Tools (5 total)

Each tool gets an optional `target` parameter (defaults to `_any`). Broadcast
targets (`_all`, labels) return per-agent results.

| Tool                  | JobClient Method                    | Parameters                                         |
| --------------------- | ----------------------------------- | -------------------------------------------------- |
| `system_hostname_get` | `QuerySystemHostname` / `Broadcast` | `target`                                           |
| `system_status_get`   | `QuerySystemStatus` / `Broadcast`   | `target`                                           |
| `network_dns_get`     | `QueryNetworkDNS` / `Broadcast`     | `target`, `interface`                              |
| `network_dns_update`  | `ModifyNetworkDNS` / `Broadcast`    | `target`, `interface`, `servers`, `search_domains` |
| `network_ping`        | `QueryNetworkPing` / `Broadcast`    | `target`, `address`                                |

### Package structure

```
internal/mcp/
├── server.go      # MCP server implementing cmd.Lifecycle
├── tools.go       # Tool registration and handlers
└── types.go       # Input structs for tools

cmd/
├── mcp.go              # mcp command group
├── mcp_server.go       # mcp server subcommand group
└── mcp_server_start.go # mcp server start (NATS + JobClient + MCP wiring)
```

### Config

Add to `internal/config/types.go`:

```go
type MCP struct {
    Server MCPServer `mapstructure:"server,omitempty"`
}

type MCPServer struct {
    Port int `mapstructure:"port"`
}
```

Add to `osapi.yaml`:

```yaml
mcp:
  server:
    port: 3001
```

## Files to create/modify

| File                                 | Change                                    |
| ------------------------------------ | ----------------------------------------- |
| `go.mod` / `go.sum`                  | Add `modelcontextprotocol/go-sdk`         |
| `internal/config/types.go`           | Add `MCP` / `MCPServer` types to `Config` |
| `osapi.yaml`                         | Add `mcp.server.port` default             |
| `internal/mcp/server.go`             | **New** — MCP server with Lifecycle       |
| `internal/mcp/tools.go`              | **New** — tool registration + handlers    |
| `internal/mcp/types.go`              | **New** — input structs                   |
| `cmd/mcp.go`                         | **New** — `mcp` command group             |
| `cmd/mcp_server.go`                  | **New** — `mcp server` subcommand         |
| `cmd/mcp_server_start.go`            | **New** — `mcp server start` command      |
| `internal/mcp/server_public_test.go` | **New** — lifecycle tests                 |
| `internal/mcp/tools_test.go`         | **New** — tool handler tests              |

## Notes

- Tool handlers follow the same pattern as REST handlers in
  `internal/api/system/` — check `job.IsBroadcastTarget()`, call the appropriate
  `JobClient` method, format results.
- stdio transport is the standard for local AI tool integration. The server runs
  as a subprocess spawned by the AI tool.
- Streamable HTTP is for remote deployments (replaces deprecated SSE).
- The NATS embedded server's `Stop()` doesn't need a context (internally
  bounded), so the adapter pattern from `cmd/nats_server_start.go` applies only
  to that server, not here.
