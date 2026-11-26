# Get

Get the metadata of a key.

The command shows the metadata of the key, and never its value.

```bash
osapi client secret get [flags]
```

## Examples

```bash
osapi client secret get --vault default --key registry-token
osapi client secret get --vault default --key registry-token --json
```

## Result

The command prints:

```ansi
$ osapi client secret get --vault default --key registry-token
[1;38;2;0;213;255mSecret:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mVault:[0m    default
  [38;2;10;143;176mKey:[0m      registry-token
  [38;2;10;143;176mVersion:[0m  3
  [38;2;10;143;176mChanged:[0m  2026-01-15 09:12:00
  [38;2;10;143;176mBy:[0m       ops@example.com
```

With `--json`, the API's response:

```json
{
  "vault": "default",
  "key": "registry-token",
  "version": 3,
  "changed": "2026-01-15T09:12:00Z",
  "by": "ops@example.com"
}
```

## Permission

It needs the `secret:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

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
