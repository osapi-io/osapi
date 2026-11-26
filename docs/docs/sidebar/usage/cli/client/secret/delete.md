# Delete

Delete a key from a vault.

Delete a key from its vault. A request that uses its reference after that fails.

```bash
osapi client secret delete [flags]
```

## Examples

```bash
osapi client secret delete --vault default --key registry-token
```

## Result

The command prints:

```ansi
$ osapi client secret delete --vault default --key registry-token
[38;2;103;234;147m✓[0m registry-token deleted

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the command prints:

```json
{ "key": "registry-token", "result": "deleted" }
```

## Permission

It needs the `secret:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag      | Description                                            | Default |
| --------- | ------------------------------------------------------ | ------- |
| `--key`   | The name of the key, such as registry-token (required) |         |
| `--vault` | The name of the vault, such as default (required)      |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
