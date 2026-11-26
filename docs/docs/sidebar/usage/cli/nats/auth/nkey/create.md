# Create

Make an nkey for each osapi role.

It makes an nkey for the setup, controller and agent roles, and writes the seed
of each (`setup.nk`, `controller.nk`, `agent.nk`), `osapi-nats.yaml`, the part
of osapi.yaml with the nkeys of the embedded server and their roles
(`nkey_roles`), and `nats-server.conf`, the authorization block of a NATS server
that osapi does not run, with the least permissions of each role. Each file is
mode 0600. A join over NATS needs a password or a JWT for its bootstrap user:
use `password create` or `operator create` for it. `--print` also prints the
server's block.

```bash
osapi nats auth nkey create [flags]
```

## Examples

```bash
osapi nats auth nkey create --out /etc/osapi/nats
```

## Result

The command prints:

```ansi
$ osapi nats auth nkey create --out /etc/osapi/nats
[1;38;2;0;213;255mNATS auth:[0m

[1;38;2;255;0;255mFiles (5):[0m
  [38;2;176;48;176mROLE        [0m[38;2;176;48;176mPATH[0m
  setup       /etc/osapi/nats/setup.nk
  controller  /etc/osapi/nats/controller.nk
  agent       /etc/osapi/nats/agent.nk
  osapi       /etc/osapi/nats/osapi-nats.yaml
  server      /etc/osapi/nats/nats-server.conf
```

With `--json`, the command prints:

```json
{
  "files": [
    {
      "role": "setup",
      "path": "/etc/osapi/nats/setup.nk"
    },
    {
      "role": "controller",
      "path": "/etc/osapi/nats/controller.nk"
    },
    {
      "role": "agent",
      "path": "/etc/osapi/nats/agent.nk"
    },
    {
      "role": "osapi",
      "path": "/etc/osapi/nats/osapi-nats.yaml"
    },
    {
      "role": "server",
      "path": "/etc/osapi/nats/nats-server.conf"
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
