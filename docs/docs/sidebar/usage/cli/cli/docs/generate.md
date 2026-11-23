# Generate

Write the CLI reference pages of the docs site.

It writes a page for each command of this binary into `--dir`, and removes each
page there that no command has. `just cli-docs-generate` runs it, and
`just test` checks that the pages match the binary.

```bash
osapi cli docs generate [flags]
```

## Examples

```bash
osapi cli docs generate
osapi cli docs generate --dir /tmp/cli
```

## Result

The command prints:

```ansi
$ osapi cli docs generate
[1;38;2;0;213;255mCLI reference:[0m
  [38;2;10;143;176mDir:[0m    docs/docs/sidebar/usage/cli
  [38;2;10;143;176mPages:[0m  171
```

With `--json`, the command prints:

```json
{
  "dir": "docs/docs/sidebar/usage/cli",
  "pages": 171
}
```

## Flags

| Flag    | Description                          | Default                       |
| ------- | ------------------------------------ | ----------------------------- |
| `--dir` | Directory of the CLI reference pages | `docs/docs/sidebar/usage/cli` |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
