# Create

Make a user and a random password for each osapi role.

It makes a user for each osapi role (setup, controller, agent and join), each
with a random password and the least permissions of its role. It writes
`osapi-nats.yaml`, the part of osapi.yaml with the users of the embedded server
and the auth of the controller and the agent, and `nats-server.conf`, the
authorization block of a NATS server that osapi does not run. Each file is mode
0600, and no password is on the command line. `--print` also prints the server's
block.

```bash
osapi nats auth password create [flags]
```

## Examples

```bash
osapi nats auth password create --out /etc/osapi/nats
```

## Result

The command prints:

```ansi
$ osapi nats auth password create --out /etc/osapi/nats
[1;38;2;0;213;255mNATS auth:[0m

[1;38;2;255;0;255mFiles (2):[0m
  [38;2;176;48;176mROLE    [0m[38;2;176;48;176mPATH[0m
  osapi   /etc/osapi/nats/osapi-nats.yaml
  server  /etc/osapi/nats/nats-server.conf
```

With `--json`, the command prints:

```json
{
  "files": [
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
