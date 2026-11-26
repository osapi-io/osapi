# Install

Install the agent as a service, with a join string, and start it.

Run this as root on the host. It writes the config file with the join string and
agent.pki.key_dir, a systemd unit (Linux) or a launch daemon (macOS) that runs
this binary as osapi agent start, and starts it. Then it waits up to a minute
for the join. After the join, the host keeps its keys and its NATS user in
agent.pki.key_dir, and each start at boot reads them: the join string is used
one time.

```bash
osapi agent service install [flags]
```

## Examples

```bash
osapi agent service install --join 'tls://nats.example.com:4222?creds=...&fp=SHA256%3A...&ns=osapi&token=osapi_jt_...'
```

## Result

The command prints:

```ansi
$ osapi agent service install --join 'tls://nats.example.com:4222?creds=...&fp=SHA256%3A...&ns=osapi&token=osapi_jt_...'
[1;38;2;0;213;255mService:[0m
  [38;2;10;143;176mService:[0m  /etc/systemd/system/osapi-agent.service
  [38;2;10;143;176mConfig:[0m   /etc/osapi/osapi.yaml
  [38;2;10;143;176mJoined:[0m   yes
  [38;2;10;143;176mNext:[0m     osapi client admission list, on a host with the CLI
```

With `--json`, the command prints:

```json
{
  "service": "/etc/systemd/system/osapi-agent.service",
  "config_file": "/etc/osapi/osapi.yaml",
  "joined": true,
  "next": "osapi client admission list, on a host with the CLI"
}
```

## Flags

| Flag       | Description                                                                | Default |
| ---------- | -------------------------------------------------------------------------- | ------- |
| `--force`  | Replace a config file that exists                                          |         |
| `--join`   | The join string that osapi controller join token create printed (required) |         |
| `--labels` | Labels of the host, as key=value, beside the labels of the join token      |         |

## Global flags

| Flag               | Description                                                                                                             | Default                 |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                                                            |                         |
| `-j, --json`       | Enable JSON output                                                                                                      |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults                                        | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`     | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
