# Get

Get the memory usage of the host.

Reads the total, the free and the used memory, in bytes. Used is the total less
what is available, which includes the cache, as free(1) reports it.

```bash
osapi client node memory get [flags]
```

## Examples

```bash
# A host with 8 GiB, 2 GiB of it free
osapi client node memory get --target web-01
```

## Permission

It needs the `memory:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
