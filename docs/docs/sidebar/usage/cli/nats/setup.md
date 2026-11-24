# Setup

Create osapi's streams and buckets on a NATS server.

It connects as the controller does, with `controller.nats`: its host, port,
namespace and auth. Run it one time before you start the controller and the
agents with that server.

Run it again after you change the settings of a stream or a bucket. Each stream
and bucket is created or updated to match the config, so running it again
changes nothing. `osapi nats start` does the same for the embedded server when
the server starts. Thus a deployment that runs the NATS server of osapi does not
need this command.

```bash
osapi nats setup
```

## Examples

```bash
osapi nats setup -f /etc/osapi/osapi.yaml
```

## Result

The command prints:

```ansi
$ osapi nats setup -f /etc/osapi/osapi.yaml
[1;38;2;0;213;255mNATS:[0m
  [38;2;10;143;176mHost:[0m       localhost
  [38;2;10;143;176mPort:[0m       4222
  [38;2;10;143;176mNamespace:[0m  osapi
  [38;2;10;143;176mAuth:[0m       none
```

With `--json`, the command prints:

```json
{
  "host": "localhost",
  "port": 4222,
  "namespace": "osapi",
  "auth": "none"
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
