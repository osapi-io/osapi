# Fingerprint

Show the fingerprint of the agent's keys, and what admits this machine.

Run this on the machine. It reads the two agent keys in `agent.pki.key_dir`
(`/etc/osapi/pki/agent` by default): the Ed25519 signing key and the X25519
encryption key. It prints the machine ID, the hostname, the public encryption
key, the fingerprint that covers the two keys, and the command that admits the
machine. The controller admits the machine only when the fingerprint is the one
of the keys in its request. The command makes each key that does not exist.

```bash
osapi agent key fingerprint
```

## Examples

```bash
osapi agent key fingerprint
```

## Result

The command prints:

```ansi
$ osapi agent key fingerprint
[1;38;2;0;213;255mAdmission:[0m
  [38;2;10;143;176mMachine ID:[0m      d530664edc1c431eb5ef1e9dfd6617a5
  [38;2;10;143;176mHostname:[0m        web-03
  [38;2;10;143;176mEncryption key:[0m  XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM
  [38;2;10;143;176mFingerprint:[0m     SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a
  [38;2;10;143;176mTo admit:[0m        osapi client admission admit --machine-id d530664edc1c431eb5ef1e9dfd6617a5 --fingerprint SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a
```

With `--json`, the command prints:

```json
{
  "machine_id": "d530664edc1c431eb5ef1e9dfd6617a5",
  "hostname": "web-03",
  "encryption_key": "XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM",
  "fingerprint": "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
  "admit": "osapi client admission admit --machine-id d530664edc1c431eb5ef1e9dfd6617a5 --fingerprint SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a"
}
```

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-j, --json`       | Enable JSON output                                                                                                      |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults                                        | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
