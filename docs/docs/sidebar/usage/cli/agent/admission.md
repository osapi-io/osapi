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

| Flag                         | Description                                                                                                              | Default                 |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ----------------------- |
| `--agent-client-name`        | NATS client name for agent                                                                                               | `osapi-agent`           |
| `--agent-host`               | NATS server hostname for agent                                                                                           | `localhost`             |
| `--agent-hostname`           | Agent hostname (defaults to system hostname)                                                                             |                         |
| `--agent-max-jobs`           | Maximum concurrent jobs per agent                                                                                        | `10`                    |
| `--agent-port`               | NATS server port for agent                                                                                               | `4222`                  |
| `--agent-queue-group`        | NATS queue group for load balancing                                                                                      | `job-agents`            |
| `--consumer-ack-wait`        | Time to wait for acknowledgment before retry                                                                             | `2m`                    |
| `--consumer-back-off`        | Retry backoff intervals                                                                                                  | `[30s,2m,5m,15m,30m]`   |
| `--consumer-max-ack-pending` | Maximum unacknowledged messages                                                                                          | `1000`                  |
| `--consumer-max-deliver`     | Maximum delivery attempts before DLQ                                                                                     | `5`                     |
| `--consumer-replay-policy`   | Replay policy: instant or original                                                                                       | `instant`               |
| `-d, --debug`                | Enable or disable debug mode                                                                                             |                         |
| `-j, --json`                 | Enable JSON output                                                                                                       |                         |
| `-f, --osapi-file`           | Path to config file                                                                                                      | `/etc/osapi/osapi.yaml` |
| `--sysctl-dir`               | Directory the agent writes osapi's kernel parameter files to, one osapi-&lt;key&gt;.conf per key. (behaviors.sysctl.dir) | `/etc/sysctl.d`         |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
