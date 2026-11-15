---
sidebar_position: 11
---

# NTP management

osapi manages the NTP (Network Time Protocol) configuration of target hosts
through chrony. The agent writes the configuration as a drop-in file under
`/etc/chrony/conf.d/`, and reloads chrony to apply the change immediately.
`create` and `update` are idempotent on the content. They calculate the hash of
the drop-in on disk, and write it again only when it is different from the
requested configuration. Thus the agent puts back a file that somebody changed
by hand.

## How it works

The NTP provider writes a single drop-in file:

```
/etc/chrony/conf.d/osapi-ntp.conf
```

The file contains `server` directives for each configured NTP server address:

```
server 0.pool.ntp.org iburst
server 1.pool.ntp.org iburst
```

After the agent writes the file, it tells chrony to reload its configuration.
Thus the change takes effect without a full restart of the daemon. The
file-state KV bucket records the SHA-256 of the deployed file. If the server
list did not change, the agent does not write the file again, and returns
`changed: false`.

The get operation runs `chronyc tracking` and `chronyc sources`. It reports the
live sync status, the stratum, the estimated offset and the current reference
source. It also reads the deployed drop-in file to report the configured server
list.

### Delete behavior

When you delete the NTP configuration, the agent removes the drop-in file
`/etc/chrony/conf.d/osapi-ntp.conf` and tells chrony to reload. Then the system
uses the other chrony configuration that stays, usually the default of the
distribution.

## Operations

| Operation | Description                                           |
| --------- | ----------------------------------------------------- |
| Get       | Get NTP sync status, stratum, offset, and server list |
| Create    | Deploy the drop-in file (idempotent on its content)   |
| Update    | Replace the drop-in file (fails if not managed)       |
| Delete    | Remove the drop-in file and reload chrony             |

## CLI usage

```bash
# Get NTP status from a host
osapi client node ntp get --target web-01

# Create NTP configuration
osapi client node ntp create --target web-01 \
  --servers 0.pool.ntp.org --servers 1.pool.ntp.org

# Update NTP servers
osapi client node ntp update --target web-01 \
  --servers ntp.example.com

# Broadcast create to all hosts
osapi client node ntp create --target _all \
  --servers 0.pool.ntp.org --servers 1.pool.ntp.org

# Delete NTP configuration
osapi client node ntp delete --target web-01
```

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All NTP operations accept broadcast targets. Use `--target _all` to apply
configuration across all registered agents. Use a label selector, for example
`--target group:web`, to target a part of the fleet.

Responses always include per-host results. The get response includes sync status
fields per host:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  SYNCHRONIZED  STRATUM  OFFSET    SOURCE          SERVERS
  web-01    yes           2        +0.000123  192.0.2.1       0.pool.ntp.org, 1.pool.ntp.org
  web-02    yes           2        +0.000045  192.0.2.1       0.pool.ntp.org, 1.pool.ntp.org
```

Skipped and failed hosts appear with their error in the output.

## Idempotent updates

Create and update compare the new server list with the SHA-256 in the file-state
KV. If the configuration did not change, the response is `changed: false`. The
agent does not write the file and does not reload chrony. You can run these
operations many times safely.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, NTP operations return `status: skipped`.
They do not fail. Refer to [Platform Detection](../sdk/platform/detection.md)
for how osapi finds the OS family.

## Permissions

| Operation              | Permission  |
| ---------------------- | ----------- |
| Get                    | `ntp:read`  |
| Create, Update, Delete | `ntp:write` |

All built-in roles (`admin`, `write`, `read`) include `ntp:read`. The `admin`
and `write` roles also include `ntp:write`.

## Related

- [CLI Reference](../usage/cli/client/node/ntp/ntp.mdx): NTP commands
- [File Management](file-management.md): file-state KV tracking
- [Platform Detection](../sdk/platform/detection.md): OS family detection
- [Configuration](../usage/configuration.md): full configuration reference
