# Create

Make an operator, an account, and a credentials file for each osapi role.

It makes an operator, its system account, and an account for osapi with
JetStream and a signing key. It writes the server's resolver file
(`resolver.conf`), a credentials file for the setup, controller and agent roles,
each with the least permissions of its role, the signing key that the controller
mints the users of a join with (`signing.nk`), the operator's seed
(`operator.nk`), and `osapi-nats.yaml`, the part of osapi.yaml that uses them.
Each file is mode 0600. Keep `operator.nk` offline: it signs the accounts.
`--print` also prints the resolver file.

```bash
osapi nats auth operator create [flags]
```

## Examples

```bash
osapi nats auth operator create --out /etc/osapi/nats
```

## Result

The command prints:

```ansi
$ osapi nats auth operator create --out /etc/osapi/nats
[1;38;2;0;213;255mNATS auth:[0m

[1;38;2;255;0;255mFiles (7):[0m
  [38;2;176;48;176mROLE        [0m[38;2;176;48;176mPATH[0m
  agent       /etc/osapi/nats/agent.creds
  controller  /etc/osapi/nats/controller.creds
  operator    /etc/osapi/nats/operator.nk
  osapi       /etc/osapi/nats/osapi-nats.yaml
  server      /etc/osapi/nats/resolver.conf
  setup       /etc/osapi/nats/setup.creds
  controller  /etc/osapi/nats/signing.nk
```

With `--json`, the command prints:

```json
{
  "files": [
    {
      "role": "agent",
      "path": "/etc/osapi/nats/agent.creds"
    },
    {
      "role": "controller",
      "path": "/etc/osapi/nats/controller.creds"
    },
    {
      "role": "operator",
      "path": "/etc/osapi/nats/operator.nk"
    },
    {
      "role": "osapi",
      "path": "/etc/osapi/nats/osapi-nats.yaml"
    },
    {
      "role": "server",
      "path": "/etc/osapi/nats/resolver.conf"
    },
    {
      "role": "setup",
      "path": "/etc/osapi/nats/setup.creds"
    },
    {
      "role": "controller",
      "path": "/etc/osapi/nats/signing.nk"
    }
  ]
}
```

## Flags

| Flag      | Description                                                                                | Default |
| --------- | ------------------------------------------------------------------------------------------ | ------- |
| `--out`   | The directory that the files go to. It is made, mode 0700, when it is not there (required) |         |
| `--print` | Also print the server's part: its authorization block, or its resolver file                |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
