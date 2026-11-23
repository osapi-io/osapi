# Get

Print a skill, or one of its reference files.

It prints the file as it is. `--json` has the name, the reference and the
content.

```bash
osapi skill get [flags]
```

## Examples

```bash
osapi skill get --name build-a-behavior
osapi skill get --name build-a-behavior --reference references/services.md
```

## Result

The command prints:

```ansi
$ osapi skill get --name build-a-behavior
---
name: build-a-behavior
description: Build or change an osapi behavior.
---

# Build a behavior

A behavior is the typed definition under internal/behaviors/<name>/.
```

With `--json`, the command prints:

```json
{
  "name": "build-a-behavior",
  "content": "---\nname: build-a-behavior\ndescription: Build or change an osapi behavior.\n---\n\n# Build a behavior\n\nA behavior is the typed definition under internal/behaviors/\u003cname\u003e/.\n"
}
```

## Flags

| Flag          | Description                                              | Default |
| ------------- | -------------------------------------------------------- | ------- |
| `--name`      | Name of the skill (required)                             |         |
| `--reference` | A reference file of the skill, relative to its directory |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
