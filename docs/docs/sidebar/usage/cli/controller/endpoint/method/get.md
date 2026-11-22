# Get

Show one method of an endpoint group: its route, fields, output, examples and
errors.

Each input field shows where the request carries it: the path, the query, a form
or the body. `--json` has the JSON Schema of the input and the output.
`osapi behavior method get` shows a method of a behavior the same way.

```bash
osapi controller endpoint method get [flags]
```

## Examples

```bash
osapi controller endpoint method get --name agent --method drain
osapi controller endpoint method get --name agent --method drain --json
```

## Flags

| Flag       | Description                                               | Default |
| ---------- | --------------------------------------------------------- | ------- |
| `--method` | The method, from osapi controller endpoint get (required) |         |
| `--name`   | The endpoint group, such as agent (required)              |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
