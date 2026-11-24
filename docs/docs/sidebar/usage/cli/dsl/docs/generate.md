# Generate

Write the CLI, DSL and reference pages of the docs site.

It writes each generated page of the docs site from this binary, and removes
each generated page that no command, DSL or group has now:

- a page for each command, under `docs/docs/sidebar/usage/cli`;
- a page for each DSL, under `docs/docs/sidebar/development/dsl`, from the spec
  that the `spec get` command of the DSL prints;
- a page for each behavior, chain, endpoint group and local group, under
  `docs/docs/sidebar/reference`, from the output of its `list` or `get` command,
  such as `osapi dsl behavior get --name dns`.

A DSL page that a person wrote stays. `just cli-docs-generate` runs it, and
`just test` checks that the pages match the binary.

```bash
osapi dsl docs generate [flags]
```

## Examples

```bash
osapi dsl docs generate
osapi dsl docs generate --root ~/src/osapi
```

## Result

The command prints:

```ansi
$ osapi dsl docs generate
[1;38;2;0;213;255mDocs:[0m
  [38;2;10;143;176mCLI pages:[0m        244
  [38;2;10;143;176mDSL pages:[0m        4
  [38;2;10;143;176mReference pages:[0m  93
```

With `--json`, the command prints:

```json
{
  "cli": 244,
  "dsl": 4,
  "reference": 93
}
```

## Flags

| Flag     | Description                  | Default |
| -------- | ---------------------------- | ------- |
| `--root` | Root of the osapi repository | `.`     |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
