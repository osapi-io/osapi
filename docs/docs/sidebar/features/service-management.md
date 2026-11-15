---
sidebar_position: 25
---

# Service management

osapi manages systemd services on managed hosts. You can list, examine, start,
stop, restart, enable and disable services. You can also create, update and
delete unit files. osapi deploys unit files through the Object Store, with the
same SHA checksums as file management.

## How it works

The service provider uses `systemctl` on Debian family systems to manage systemd
service units.

### List

Returns all systemd services on the host. Each entry has the service name, the
active status, the enabled state, the description and the PID, if the service
runs.

### Get

Returns the details of one service: its name, active status, enabled state,
description and PID.

### Start

Starts a stopped service with `systemctl start`. If the service was not running,
it returns `changed: true`. If the service was already active, it returns
`changed: false`.

### Stop

Stops a running service with `systemctl stop`. If the service was running, it
returns `changed: true`. If the service was already stopped, it returns
`changed: false`.

### Restart

Restarts a service with `systemctl restart`. It always returns `changed: true`,
because the agent restarts the service in all states.

### Enable

Enables a service with `systemctl enable`, so that it starts at boot. If the
service was not enabled, it returns `changed: true`.

### Disable

Disables a service with `systemctl disable`, so that it does not start at boot.
If the service was enabled, it returns `changed: true`.

### Create

Deploys a new service unit file to the host. First, upload the content of the
unit file to the Object Store. The agent writes the file to
`/etc/systemd/system/{name}` and runs `systemctl daemon-reload` to load the new
unit.

Create is idempotent **on the content**. If a unit file on disk has the same
content, it reports `changed: false`. If somebody changed the file, the agent
writes it again from the Object Store. Use `update` to point the host to a
different object.

### Update

Replaces a service unit file with a new Object Store reference. The agent
deploys the unit file again and runs `systemctl daemon-reload`. If the content
did not change (the same SHA), it returns `changed: false` and does not reload
the daemon. It fails if the unit file does not exist. Use `create` first.

### Delete

Removes a service unit file from the host. If the service runs, the agent stops
and disables it. Then the agent deletes the unit file from
`/etc/systemd/system/` and runs `systemctl daemon-reload`.

## Operations

| Operation | Description                                          |
| --------- | ---------------------------------------------------- |
| List      | List all systemd services                            |
| Get       | Get the details of one service                       |
| Start     | Start a service                                      |
| Stop      | Stop a service                                       |
| Restart   | Restart a service                                    |
| Enable    | Enable a service, so that it starts at boot          |
| Disable   | Disable a service, so that it does not start at boot |
| Create    | Deploy a service unit file from Object Store         |
| Update    | Deploy a service unit file again                     |
| Delete    | Remove a service unit file and stop the unit         |

## CLI usage

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

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All service operations accept broadcast targets. Use `--target _all` to manage
the services on all registered agents. Use a label selector, for example
`--target group:web`, to target a part of the fleet.

Responses always have a result for each host:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    nginx.service    true
  web-02    nginx.service    true
```

The output shows skipped and failed hosts with their error.

## Supported platforms

| OS family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, service operations return
`status: skipped`. They do not fail. Refer to
[Platform Detection](../sdk/platform/detection.md) for how osapi finds the OS
family.

## Container behavior

In a container, service operations return `status: skipped`. The provider
returns `ErrUnsupported`, because `systemctl` needs systemd as PID 1. A standard
container does not have it.

## Permissions

| Operation                                                     | Permission      |
| ------------------------------------------------------------- | --------------- |
| List, Get                                                     | `service:read`  |
| Start, Stop, Restart, Enable, Disable, Create, Update, Delete | `service:write` |

To list and examine services, you need `service:read`. All built-in roles
(`admin`, `write`, `read`) have it. The operations that change services need
`service:write`. The `admin` and `write` roles have it.

## Related

- [CLI Reference](../usage/cli/client/node/service/service.mdx): the service
  commands
- [SDK Reference](../sdk/client/services/service.md): the service of the SDK
- [Platform Detection](../sdk/platform/detection.md): how osapi finds the OS
  family
- [Configuration](../usage/configuration.md): all configuration settings
