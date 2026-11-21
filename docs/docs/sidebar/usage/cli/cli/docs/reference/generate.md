# Generate

Write the reference pages of the docs site.

It writes a page for each behavior, chain, endpoint group and local group of
this binary under `docs/docs/sidebar/reference`. Each page is the output of the
`list` or `get` command that shows it, such as `osapi behavior get --name dns`,
printed as markdown. Thus a page shows what its command shows. It removes each
generated page there that no command writes now. `just cli-docs-generate` runs
it, and `just test` checks that the pages match the binary.

```bash
osapi cli docs reference generate [flags]
```

## Examples

```bash
osapi cli docs reference generate
osapi cli docs reference generate --root ~/src/osapi
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

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
