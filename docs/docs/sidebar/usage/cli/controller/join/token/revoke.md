# Revoke

Revoke a join token, so it admits no agent.

```bash
osapi controller join token revoke [flags]
```

## Examples

```bash
osapi controller join token revoke --id 1a2b3c4d
```

## Result

The command prints:

```ansi
$ osapi controller join token revoke --id 1a2b3c4d
[1;38;2;0;213;255mJoin token:[0m
  [38;2;10;143;176mID:[0m  1a2b3c4d
```

With `--json`, the command prints:

```json
{
  "id": "1a2b3c4d"
}
```

## Flags

| Flag   | Description                                  | Default |
| ------ | -------------------------------------------- | ------- |
| `--id` | The token, as token list shows it (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
