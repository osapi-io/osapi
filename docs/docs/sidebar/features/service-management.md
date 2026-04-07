---
sidebar_position: 25
---

# Service Management

OSAPI provides management of systemd services on managed hosts. Services can be
listed, inspected, started, stopped, restarted, enabled, disabled, and managed
through unit file CRUD operations. Unit files are deployed via the Object Store
using the same SHA-based idempotency as file management.

## How It Works

The service provider interacts with `systemctl` on Debian-family systems to
manage systemd service units.

### List

Returns all systemd services on the host. Each entry includes the service name,
active status, enabled state, description, and PID (if running).

### Get

Returns detailed information about a specific service, including its name,
active status, enabled state, description, and PID.

### Start

Starts a stopped service using `systemctl start`. Returns `changed: true` if the
service was not already running, or `changed: false` if it was already active.

### Stop

Stops a running service using `systemctl stop`. Returns `changed: true` if the
service was running, or `changed: false` if it was already stopped.

### Restart

Restarts a service using `systemctl restart`. Always returns `changed: true`
because the service is restarted regardless of its current state.

### Enable

Enables a service to start on boot using `systemctl enable`. Returns
`changed: true` if the service was not already enabled.

### Disable

Disables a service from starting on boot using `systemctl disable`. Returns
`changed: true` if the service was previously enabled.

### Create

Deploys a new service unit file to the host. The unit file content must first be
uploaded to the Object Store. The agent writes the file to
`/etc/systemd/system/{name}` and runs `systemctl daemon-reload` to pick up the
new unit. Idempotent: returns `changed: false` if already managed. Use `update`
to replace it.

### Update

Replaces an existing service unit file with a new Object Store reference. The
agent redeploys the unit file and runs `systemctl daemon-reload`. If the content
has not changed (same SHA), `changed: false` is returned and the daemon is not
reloaded. Fails if the unit file does not exist -- use `create` first.

### Delete

Removes a service unit file from the host. The agent stops and disables the
service if it is running, deletes the unit file from `/etc/systemd/system/`, and
runs `systemctl daemon-reload`.

## Operations

| Operation | Description                                  |
| --------- | -------------------------------------------- |
| List      | List all systemd services                    |
| Get       | Get details for a specific service           |
| Start     | Start a service                              |
| Stop      | Stop a service                               |
| Restart   | Restart a service                            |
| Enable    | Enable a service to start on boot            |
| Disable   | Disable a service from starting on boot      |
| Create    | Deploy a service unit file from Object Store |
| Update    | Redeploy an existing service unit file       |
| Delete    | Remove a service unit file and stop the unit |

## CLI Usage

```bash
# List all services on a host
osapi client node service list --target web-01

# Get details for a specific service
osapi client node service get --target web-01 --name nginx.service

# Start a service
osapi client node service start --target web-01 --name nginx.service

# Stop a service
osapi client node service stop --target web-01 --name nginx.service

# Restart a service
osapi client node service restart --target web-01 --name nginx.service

# Enable a service to start on boot
osapi client node service enable --target web-01 --name nginx.service

# Disable a service from starting on boot
osapi client node service disable --target web-01 --name nginx.service

# Deploy a new unit file from the Object Store
osapi client node service create --target web-01 \
  --name myapp.service --object myapp-unit

# Update an existing unit file
osapi client node service update --target web-01 \
  --name myapp.service --object myapp-unit-v2

# Delete a unit file
osapi client node service delete --target web-01 \
  --name myapp.service

# Broadcast service restart to all hosts
osapi client node service restart --target _all --name nginx.service
```

All commands support `--json` for raw JSON output.

## Broadcast Support

All service operations support broadcast targeting. Use `--target _all` to
manage services on every registered agent, or use a label selector like
`--target group:web` to target a subset.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    nginx.service    true
  web-02    nginx.service    true
```

Skipped and failed hosts appear with their error in the output.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, service operations return `status: skipped` instead of
failing. See [Platform Detection](../sdk/platform/detection.md) for details on
OS family detection.

## Container Behavior

Service operations return `status: skipped` inside containers. The provider
returns `ErrUnsupported` because `systemctl` requires systemd as PID 1, which is
not available in standard container environments.

## Permissions

| Operation                                                     | Permission      |
| ------------------------------------------------------------- | --------------- |
| List, Get                                                     | `service:read`  |
| Start, Stop, Restart, Enable, Disable, Create, Update, Delete | `service:write` |

Service listing and inspection require `service:read`, included in all built-in
roles (`admin`, `write`, `read`). Mutation and action operations require
`service:write`, included in the `admin` and `write` roles.

## Related

- [CLI Reference](../usage/cli/client/node/service/service.md) -- service
  commands
- [SDK Reference](../sdk/client/services/service.md) -- Service service
- [Platform Detection](../sdk/platform/detection.md) -- OS family detection
- [Configuration](../usage/configuration.md) -- full configuration reference
