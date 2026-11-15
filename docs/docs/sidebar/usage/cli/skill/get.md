# Get

Print a skill, or one of its reference files.

```bash
osapi skill get [flags]
```

## Examples

```bash
osapi skill get --name build-a-behavior
osapi skill get --name build-a-behavior --reference references/services.md
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
