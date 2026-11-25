# Get

Print the whole config, each key at its default, as YAML.

Each key has its doc above it. A key with no default is a choice of the
operator, commented out with why. Keep only the keys that you change: a config
file needs no other. The defaults come from the definition that owns each
section, such as controller start, so this is never out of date.

```bash
osapi config default get
```

## Examples

```bash
osapi config default get
osapi config default get > osapi.yaml
```

## Result

The command prints:

```ansi
$ osapi config default get
controller:
  api:
    # Port the REST API server listens on.
    port: 8080
```

With `--json`, the command prints:

```json
{
  "content": "controller:\n  api:\n    # Port the REST API server listens on.\n    port: 8080\n"
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
