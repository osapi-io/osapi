# Start

Start the controller.

The controller serves the REST API, writes its heartbeat, and watches conditions
to send notifications.

```bash
osapi controller start [flags]
```

## Examples

```bash
osapi controller start
osapi controller start --controller-api-port 9000
```

## Flags

| Flag                    | Description                                               | Default |
| ----------------------- | --------------------------------------------------------- | ------- |
| `--controller-api-port` | Port the REST API server listens on (controller.api.port) | `8080`  |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
