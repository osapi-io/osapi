# Drain

Drain an agent.

The agent changes from `Ready` to `Draining`. When all its running jobs are
complete, the state is `Cordoned`. The agent continues to run and to send
heartbeats. It only stops the intake of new work from the job queue. Use
`agent undrain` to accept jobs again. The jobs that it runs now continue until
they are complete.

```bash
osapi client agent drain [flags]
```

## Examples

```bash
osapi client agent drain --hostname web-01
```

## Result

The command prints:

```ansi
$ osapi client agent drain --hostname web-01
[38;2;103;234;147m✓[0m web-01 draining

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the command prints:

```json
{ "hostname": "web-01", "result": "draining" }
```

## Permission

It needs the `agent:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag         | Description                                   | Default |
| ------------ | --------------------------------------------- | ------- |
| `--hostname` | The hostname of the agent to drain (required) |         |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
