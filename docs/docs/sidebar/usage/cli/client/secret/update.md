# Update

Write the value of a key.

The command reads the value from standard input with --value -, so the value is
not in the shell history. A request then uses the key as
secret://&lt;vault&gt;/&lt;key&gt;.

```bash
osapi client secret update [flags]
```

## Examples

```bash
# Read the value from standard input
osapi client secret update --vault default --key registry-token --value -
```

## Result

Read the value from standard input:

```ansi
$ osapi client secret update --vault default --key registry-token --value -
[1;38;2;0;213;255mSecret:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mVault:[0m    default
  [38;2;10;143;176mKey:[0m      registry-token
  [38;2;10;143;176mVersion:[0m  4
  [38;2;10;143;176mChanged:[0m  2026-01-15 10:00:00
  [38;2;10;143;176mBy:[0m       ops@example.com
```

With `--json`, the API's response:

```json
{
  "vault": "default",
  "key": "registry-token",
  "version": 4,
  "changed": "2026-01-15T10:00:00Z",
  "by": "ops@example.com"
}
```

## Permission

It needs the `secret:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag      | Description                                                                                                                                     | Default |
| --------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `--key`   | The name of the key, such as registry-token (required)                                                                                          |         |
| `--value` | The value of the key, which no method gives back (a secret: secret://&lt;vault&gt;/&lt;key&gt;, or - to read it from standard input) (required) |         |
| `--vault` | The name of the vault, such as default (required)                                                                                               |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
