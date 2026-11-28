# Uninstall

Stop the agent's service and remove it.

Run this as root on the host. It stops the service and removes its unit or
launch daemon. With --purge it also removes the config file and the keys in
agent.pki.key_dir: then the host is new to osapi. To join it again, delete its
admission with osapi client admission delete, then give it a new join string.
The controller refuses a join that replaces the keys of an admitted machine.

```bash
osapi agent service uninstall [flags]
```

## Examples

```bash
osapi agent service uninstall --purge
```

## Result

The command prints:

```ansi
$ osapi agent service uninstall --purge
[1;38;2;0;213;255mRemoved:[0m
  [38;2;10;143;176mRemoved:[0m  /etc/systemd/system/osapi-agent.service
```

With `--json`, the command prints:

```json
{
  "removed": ["/etc/systemd/system/osapi-agent.service"]
}
```

## Flags

| Flag      | Description                                                   | Default |
| --------- | ------------------------------------------------------------- | ------- |
| `--purge` | Also remove the config file and the keys of agent.pki.key_dir |         |

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-j, --json`       | Enable JSON output                                                                                                      |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults                                        | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
