# Get

Print each part of a collector, what it does and when to use it.

```bash
osapi agent collector spec get [flags]
```

## Examples

```bash
osapi agent collector spec get
osapi agent collector spec get --json
```

## Flags

| Flag     | Description | Default |
| -------- | ----------- | ------- |
| `--json` | Print JSON  |         |

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults                                        | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
