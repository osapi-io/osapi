# List

List the collectors: category, whether it runs by default, dependencies.

```bash
osapi agent collector list [flags]
```

## Examples

```bash
osapi agent collector list
osapi agent collector list --json
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
