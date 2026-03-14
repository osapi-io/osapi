---
sidebar_position: 1
---

# Client Library

The `osapi` package provides a typed Go client for the OSAPI REST API. Create a
client with `New()` and use domain-specific services to interact with the API.

## Quick Start

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

client := client.New("http://localhost:8080", "your-jwt-token")

resp, err := client.Node.Hostname(ctx, "_any")
```

## Services

| Service               | Description                        |
| --------------------- | ---------------------------------- |
| [Agent](agent.md)     | Agent discovery and details        |
| [Audit](audit.md)     | Audit log operations               |
| [Docker](docker.md)   | Container runtime operations       |
| [File](file.md)       | File management (Object Store)     |
| [Health](health.md)   | Health check operations            |
| [Job](job.md)         | Async job queue operations         |
| [Metrics](metrics.md) | Prometheus metrics access          |
| [Node](node.md)       | Node management, network, commands |

## Client Options

| Option                         | Description                    |
| ------------------------------ | ------------------------------ |
| `WithLogger(logger)`           | Set custom `slog.Logger`       |
| `WithHTTPTransport(transport)` | Set custom `http.RoundTripper` |

`WithLogger` defaults to `slog.Default()`. `WithHTTPTransport` sets the base
transport for HTTP requests.

## Targeting

Most operations accept a `target` parameter:

| Target      | Behavior                                    |
| ----------- | ------------------------------------------- |
| `_any`      | Send to any available agent (load balanced) |
| `_all`      | Broadcast to every agent                    |
| `hostname`  | Send to a specific host                     |
| `key:value` | Send to agents matching a label             |
