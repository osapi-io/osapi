# Remove

Remove an enrolled agent.

When PKI is enabled, the controller keeps the public key of each accepted agent.
This command deletes it. After that, the controller does not verify the
responses or the registration of the agent. Thus a target does not find the
agent, and a broadcast gives it no row.

```bash
osapi client agent remove [flags]
```

## Examples

```bash
osapi client agent remove --hostname web-01
osapi client agent remove --hostname web-01 --machine-id 3f2a9c...
```

## Permission

It needs the `agent:write` permission. `osapi token roles` lists the roles that
grant it.

## Flags

| Flag           | Description                                                                   | Default |
| -------------- | ----------------------------------------------------------------------------- | ------- |
| `--hostname`   | The hostname of the accepted agent to remove. (required)                      |         |
| `--machine-id` | The machine ID to remove, when more than one accepted agent has the hostname. |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
