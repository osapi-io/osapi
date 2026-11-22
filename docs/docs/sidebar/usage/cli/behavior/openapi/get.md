# Get

Print a served behavior's part of the OpenAPI spec, as YAML.

It prints the spec as the definition makes it: the part that go generate writes
to the behavior's gen/api.yaml, and that the join puts in the spec of the API.
An internal behavior serves no route, so it has none.

```bash
osapi behavior openapi get [flags]
```

## Examples

```bash
osapi behavior openapi get --name sysctl
```

## Flags

| Flag     | Description                                       | Default |
| -------- | ------------------------------------------------- | ------- |
| `--name` | The behavior, from osapi behavior list (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
