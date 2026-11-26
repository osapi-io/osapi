# Create

Make an operator, an account, and a credentials file for each osapi role.

It makes an operator, its system account, and an account for osapi with
JetStream. It writes the server's resolver file (`resolver.conf`), a credentials
file for the setup, controller and agent roles, each with the least permissions
of its role, and `osapi-nats.yaml`, the part of osapi.yaml that uses them. Each
file is mode 0600.

The controller mints the users of a join with two scoped signing keys:
`agent-signing.nk`, whose template is the agent role's permissions, and
`join-signing.nk`, the join role's. Thus the controller cannot mint a user with
more than an agent or a join user has. Keep `operator.nk` and `account.nk`
offline: they sign the accounts and the static users. `--print` also prints the
resolver file.

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

[1;38;2;255;0;255mFiles (9):[0m
  [38;2;176;48;176mROLE        [0m[38;2;176;48;176mPATH[0m
  agent       /etc/osapi/nats/agent.creds
  controller  /etc/osapi/nats/controller.creds
  operator    /etc/osapi/nats/operator.nk
  osapi       /etc/osapi/nats/osapi-nats.yaml
  server      /etc/osapi/nats/resolver.conf
  setup       /etc/osapi/nats/setup.creds
  controller  /etc/osapi/nats/agent-signing.nk
  controller  /etc/osapi/nats/join-signing.nk
  operator    /etc/osapi/nats/account.nk
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
      "path": "/etc/osapi/nats/agent-signing.nk"
    },
    {
      "role": "controller",
      "path": "/etc/osapi/nats/join-signing.nk"
    },
    {
      "role": "operator",
      "path": "/etc/osapi/nats/account.nk"
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
