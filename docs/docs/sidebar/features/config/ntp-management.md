# NTP management

osapi manages the NTP servers that chrony uses on target hosts. The `ntp`
behavior writes the servers to one file under `/etc/chrony/sources.d/`, and
reloads chrony's sources to apply the change at once. The behavior writes
through the `managedfile` building block. Thus a file that has the same servers
does not change, and `create` and `update` report `changed: false`.

## How it works

The behavior writes a single file:

```
/etc/chrony/sources.d/osapi-ntp.sources
```

The file has a `server` line for each server:

```
server 0.pool.ntp.org iburst
server 1.pool.ntp.org iburst
```

After the file changes, the agent runs `chronyc reload sources`. Thus chrony
reads the file without a restart. A failed reload is a warning: the file is
written, and chrony reads it at its next start. `managedfile` records the file
in the file-state KV bucket, and `update` fails with not found when that record
is missing.

The get operation runs `chronyc tracking` and `chronyc -c sources`. It reports
whether chrony is synchronized, its stratum, its offset from NTP time, the
source that it follows, and each source that it has. This is the live state of
chrony, not a fact of the host, so it does not go through gohai.

A server must be a hostname or an IP address. A server with a space or a line
break is refused with 400, so a server cannot add a directive of its own to the
file.

### Delete behavior

Delete removes `/etc/chrony/sources.d/osapi-ntp.sources` and reloads the
sources. Then chrony uses its other sources, usually the default of the
distribution. When osapi set no servers, delete reports `changed: false`.

## Operations

| Operation | Description                                           |
| --------- | ----------------------------------------------------- |
| Get       | Get NTP sync status, stratum, offset, and server list |
| Create    | Write the servers file (idempotent on its content)    |
| Update    | Replace the servers file (fails if osapi set none)    |
| Delete    | Remove the servers file and reload the sources        |

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

Create and update compare the new servers file with the file on the host. If the
file has the same content, the response is `changed: false`. The agent does not
write the file and does not reload chrony. You can run these operations many
times safely.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, and in a container, whose clock is its
host's, NTP operations return `status: skipped`. They do not fail. Refer to
[Platform Detection](../../sdk/platform/detection.md) for how osapi finds the OS
family.

## Permissions

The [ntp](../../reference/behaviors/config/ntp/ntp.md) reference page gives the
permission of each method.

## Related

- [CLI Reference](../../usage/cli/client/node/ntp/ntp.mdx): NTP commands
- [File Management](file-management.md): file-state KV tracking
- [Platform Detection](../../sdk/platform/detection.md): OS family detection
- [Configuration](../../usage/configuration.md): full configuration reference
