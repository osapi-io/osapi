# Validate

Check every collector against the rules, and print the fix of each problem.

It exits 1 when a collector breaks a rule.

The rules:

name: Each collector has a name of lowercase letters, digits and underscores,
and no two collectors have one name. category: Each collector's category is one
of the categories of internal/collector. facts: Each collector has a field of
gohai.Facts whose JSON name is its name, so the schema and the docs have its
facts. dependency: Each dependency of a collector is another registered
collector.

```bash
osapi agent collector validate [flags]
```

## Examples

```bash
osapi agent collector validate
osapi agent collector validate --json
```

## Flags

| Flag     | Description | Default |
| -------- | ----------- | ------- |
| `--json` | Print JSON  |         |

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-f, --osapi-file` | Path to config file                                                                                                     | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
