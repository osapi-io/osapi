# List

List the skills that this binary has.

```bash
osapi skill list
```

## Examples

```bash
osapi skill list
osapi skill list --json
```

## Result

The command prints:

```ansi
$ osapi skill list
[1;38;2;0;213;255mSkills (3):[0m
  [38;2;10;143;176mNAME              [0m[38;2;10;143;176mINSTALLED BY INIT[0m
  [1mbuild-a-behavior[0m  no
  [1mosapi[0m             yes
  [1mosapi-operations[0m  yes
```

With `--json`, the command prints:

```json
{
  "items": [
    {
      "name": "build-a-behavior",
      "description": "Build or change an osapi dsl behavior.",
      "init": false,
      "references": ["references/services.md"]
    },
    {
      "name": "osapi",
      "description": "Operate a fleet of hosts through osapi.",
      "init": true,
      "references": ["references/behaviors.md", "references/targets.md"]
    },
    {
      "name": "osapi-operations",
      "description": "Run and configure the components of osapi.",
      "init": true,
      "references": ["references/config.md", "references/tokens.md"]
    }
  ]
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
