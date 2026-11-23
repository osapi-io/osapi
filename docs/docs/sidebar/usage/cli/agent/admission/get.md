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

## Result

The command prints:

```ansi
$ osapi agent admission get
[1;38;2;0;213;255mAdmission:[0m
  [38;2;10;143;176mMachine ID:[0m   d530664edc1c431eb5ef1e9dfd6617a5
  [38;2;10;143;176mHostname:[0m     web-03
  [38;2;10;143;176mFingerprint:[0m  SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a
  [38;2;10;143;176mTo admit:[0m     osapi client admission admit --machine-id d530664edc1c431eb5ef1e9dfd6617a5 --fingerprint SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a
```

With `--json`, the command prints:

```json
{
  "machine_id": "d530664edc1c431eb5ef1e9dfd6617a5",
  "hostname": "web-03",
  "fingerprint": "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
  "admit": "osapi client admission admit --machine-id d530664edc1c431eb5ef1e9dfd6617a5 --fingerprint SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a"
}
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
