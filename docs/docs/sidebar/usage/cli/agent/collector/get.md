# Get

Show a collector: its constructor, its fields, and what it reads on each
platform.

The reads are what the collector reads on an empty host of each platform: the
files that it opens or looks at, and the commands that it runs. With content, it
also reads under these paths. A fake host of a test gives each one, and a fact
that no collector reads is a change to gohai first.

```bash
osapi agent collector get [flags]
```

## Examples

```bash
osapi agent collector get --name services
osapi agent collector get --name services --json
```

## Flags

| Flag     | Description                                               | Default |
| -------- | --------------------------------------------------------- | ------- |
| `--json` | Print JSON                                                |         |
| `--name` | The collector, from osapi agent collector list (required) |         |

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults                                        | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
