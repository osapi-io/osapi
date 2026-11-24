# Create

Make the key that encrypts the JetStream store.

The command writes 32 random bytes as hex to the file that
`nats.server.jetstream.key_file` names, with mode 0600. It makes the directory
of the file with mode 0700. The embedded NATS server reads the key when it
starts, and encrypts each stream and bucket in `store_dir`.

The command refuses to write a file that exists, because a new key in place of
the old key makes each stream unreadable. Keep a copy of the key apart from the
backup of `store_dir`: without the key, the server cannot read the store. Refer
to \[Configuration](../../../configuration.md).

```bash
osapi nats key create
```

## Examples

```bash
osapi nats key create -f /etc/osapi/osapi.yaml
```

## Result

The command prints:

```ansi
$ osapi nats key create -f /etc/osapi/osapi.yaml
[1;38;2;0;213;255mJetStream key:[0m
  [38;2;10;143;176mKey file:[0m  /etc/osapi/nats/jetstream.key
```

With `--json`, the command prints:

```json
{
  "key_file": "/etc/osapi/nats/jetstream.key"
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
