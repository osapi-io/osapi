# Generate

Write the DSL reference pages of the docs site.

It writes a page for each DSL of this binary into `--dir`: the behavior DSL,
chains, the endpoint DSL and the command DSL. Each page lists what the
`spec get` command of the DSL prints, from the same spec. It removes each
generated page there that no DSL has, and keeps a page that a person wrote.
`just cli-docs-generate` runs it, and `just test` checks that the pages match
the binary.

```bash
osapi cli docs dsl generate [flags]
```

## Examples

```bash
osapi cli docs dsl generate
osapi cli docs dsl generate --dir /tmp/dsl
```

## Result

The command prints:

```ansi
$ osapi cli docs dsl generate
[1;38;2;0;213;255mDSL reference:[0m
  [38;2;10;143;176mDir:[0m    docs/docs/sidebar/development/dsl
  [38;2;10;143;176mPages:[0m  7
```

With `--json`, the command prints:

```json
{
  "dir": "docs/docs/sidebar/development/dsl",
  "pages": 7
}
```

## Flags

| Flag    | Description                | Default                             |
| ------- | -------------------------- | ----------------------------------- |
| `--dir` | Directory of the DSL pages | `docs/docs/sidebar/development/dsl` |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
