# Get

Show one method of a local group: its command, flags, output, examples and
refusals.

Each flag has the validate tag of its field. `--json` has the JSON Schema of the
output. `osapi behavior method get` and `osapi controller endpoint method get`
show a method of a behavior and of an endpoint group the same way.

```bash
osapi cli command method get [flags]
```

## Examples

```bash
osapi cli command method get --name token --method generate
osapi cli command method get --name token --method generate --json
```

## Flags

| Flag       | Description                                                                  | Default |
| ---------- | ---------------------------------------------------------------------------- | ------- |
| `--method` | The method, such as list or roles get, from osapi cli command get (required) |         |
| `--name`   | The command path of the local group, such as token (required)                |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
