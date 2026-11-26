# Liveness

Liveness probe.

Returns 200 if the process is alive. No checks performed.

```bash
osapi client health liveness
```

## Examples

```bash
osapi client health liveness
osapi client health liveness --json
```

## Result

The command prints:

```ansi
$ osapi client health liveness
[1;38;2;0;213;255mHealth:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mStatus:[0m  [38;2;103;234;147mok[0m
```

With `--json`, the API's response:

```json
{
  "status": "ok"
}
```

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
