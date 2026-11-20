# Admission

Show what admits this machine.

Run this on the machine. It prints the machine ID, the hostname and the
fingerprint of the agent key, from the PKI key directory (`/etc/osapi/pki` by
default). Give the fingerprint to `osapi client admission admit`. The controller
admits the machine only when the fingerprint is the one of the key in its
request. If no key exists, the command makes a new key pair.

```bash
osapi agent admission
```

## Examples

```bash
osapi agent admission
```

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-j, --json`       | Enable JSON output                                                                                                      |                         |
| `-f, --osapi-file` | Path to config file                                                                                                     | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
