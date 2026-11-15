---
sidebar_position: 10
---

# Sysctl management

osapi manages Linux kernel parameters on target hosts with `sysctl`. The agent
writes each parameter to a drop-in file under `/etc/sysctl.d/` and applies it
immediately with `sysctl -p`. Thus a change takes effect immediately, and it
stays after a reboot.

## How it works

The sysctl provider writes one file per parameter:

```
/etc/sysctl.d/osapi-{key}.conf
```

For example, setting `net.ipv4.ip_forward` writes:

```
/etc/sysctl.d/osapi-net.ipv4.ip_forward.conf
```

File content is a single `key = value` line:

```
net.ipv4.ip_forward = 1
```

After the provider writes the file, it runs
`sysctl -p /etc/sysctl.d/osapi-{key}.conf` to apply the value to the running
kernel immediately. If `sysctl -p` fails, the operation fails, because the
kernel does not have the value.

Updates are idempotent. The provider calculates the hash of the drop-in **on
disk**, and writes it again only when it is different. When it is the same, it
returns `changed: false`. A drop-in that somebody changed by hand is a
difference, and the provider puts it back. The comparison reads the file, not
the file-state KV record of the last deploy.

The write is atomic. The provider writes a temporary file next to the drop-in
and renames it over the drop-in. Thus `sysctl --system` never reads a setting
that is only partly written.

### Delete behavior

When you delete a sysctl entry, the agent removes the drop-in file
`/etc/sysctl.d/osapi-{key}.conf`. The running kernel keeps the value. It stays
until the next reboot, or until you set a new value. To go back to the default,
set the default value before you delete the file.

### List and get

List and get read the kernel, not osapi's records. List returns every parameter
under `/proc/sys`, as `sysctl -a` does, and get returns any parameter's current
value, whoever set it. A host has about a thousand parameters.

sysctl is a [behavior](../behaviors/sysctl/sysctl.md), served under
`osapi client next node sysctl` and `/api/next` beside these commands until its
cutover. There, each entry also says whether osapi manages it, and list takes a
`--prefix`, such as `net.ipv4`, to return one part of the tree.

## Operations

| Operation | Description                                                 |
| --------- | ----------------------------------------------------------- |
| List      | List every parameter and its current value                  |
| Get       | Get any parameter's current value                           |
| Create    | Write a drop-in and apply it (idempotent)                   |
| Update    | The same as create, addressed by key; manages it if not yet |
| Delete    | Remove the drop-in; the value stays until reboot            |

## CLI usage

```bash
# List every parameter on a host
osapi client node sysctl list --target web-01

# Get a specific parameter
osapi client node sysctl get --target web-01 \
  --key net.ipv4.ip_forward

# Set a parameter and manage it (idempotent: unchanged if already set)
osapi client node sysctl create --target web-01 \
  --key net.ipv4.ip_forward --value 1

# Set it by key, whether or not osapi managed it before
osapi client node sysctl update --target web-01 \
  --key net.ipv4.ip_forward --value 0

# Broadcast create to all hosts
osapi client node sysctl create --target _all \
  --key vm.swappiness --value 10

# Delete a parameter (removes the drop-in file)
osapi client node sysctl delete --target web-01 \
  --key net.ipv4.ip_forward
```

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All sysctl operations accept broadcast targets. Use `--target _all` to apply a
parameter across all registered agents. Use a label selector, for example
`--target group:web`, to target a part of the fleet.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  KEY                   VALUE
  web-01    net.ipv4.ip_forward   1
  web-02    net.ipv4.ip_forward   1
```

Skipped and failed hosts appear with their error in the output.

## Idempotent updates

Create and update compare the drop-in on disk with the content to write. If the
drop-in already has the value, the response is `changed: false`. You can run
these operations many times safely. They do not write files or run `sysctl -p`
without a need.

## Key naming

Sysctl keys use the standard dot notation of the Linux kernel, for example
`net.ipv4.ip_forward`, `vm.swappiness` and `kernel.hostname`. osapi accepts keys
with letters, digits, dots, underscores and hyphens. A key must start with a
letter or a digit (`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`). This includes each key that
sysctl knows. The API and the provider of the agent both validate the shape of
the key and the value before they apply a change.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, sysctl operations return `status: skipped`.
They do not fail. Refer to [Platform Detection](../sdk/platform/detection.md)
for how osapi finds the OS family.

## Permissions

| Operation              | Permission     |
| ---------------------- | -------------- |
| List, Get              | `sysctl:read`  |
| Create, Update, Delete | `sysctl:write` |

All built-in roles (`admin`, `write`, `read`) include `sysctl:read`. The `admin`
and `write` roles also include `sysctl:write`.

## Related

- [CLI Reference](../usage/cli/client/node/sysctl/sysctl.mdx): sysctl commands
- [File Management](file-management.md): file-state KV tracking
- [Platform Detection](../sdk/platform/detection.md): OS family detection
- [Configuration](../usage/configuration.md): full configuration reference
