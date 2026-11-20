# Get

Show what admits this machine.

Run this on the machine. It prints the machine ID, the hostname and the
fingerprint of the agent key, in `agent.pki.key_dir` (`/etc/osapi/pki` by
default), and the command that admits the machine. The controller admits the
machine only when the fingerprint is the one of the key in its request. If no
key exists, the command makes a new key pair.

```bash
osapi agent admission get
```

## Examples

```bash
osapi agent admission get
```

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-j, --json`       | Enable JSON output                                                                                                      |                         |
| `-f, --osapi-file` | Path to config file                                                                                                     | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
