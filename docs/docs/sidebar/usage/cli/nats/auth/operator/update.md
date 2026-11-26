# Update

Sign the account and each role's credentials again, with the same keys.

It reads the keys that `operator create` wrote to the directory, and writes
`resolver.conf` and the credentials file of each osapi role again, with the
permissions that each role has in this version of osapi. Run it after an upgrade
that adds a permission to a role, such as the rotation of an agent's keys. The
account's JWT holds the agent's permissions as the template of its scoped
signing key, and each credentials file holds the permissions of its role, so an
upgrade does not change them. The keys stay, so each agent keeps its user and no
agent joins again. Restart the NATS server, which reads the resolver file at its
start, then the controller, for its new credentials. It needs `operator.nk` and
`account.nk`.

```bash
osapi nats auth operator update [flags]
```

## Examples

```bash
osapi nats auth operator update --out /etc/osapi/nats
```

## Result

The command prints:

```ansi
$ osapi nats auth operator update --out /etc/osapi/nats
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

| Flag      | Description                                                                                                  | Default |
| --------- | ------------------------------------------------------------------------------------------------------------ | ------- |
| `--out`   | The directory that operator create wrote: update reads its keys there, and writes the files again (required) |         |
| `--print` | Also print the server's resolver file                                                                        |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
