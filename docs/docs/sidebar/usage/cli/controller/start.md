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

## Result

The command prints:

```ansi
$ osapi controller start
[2m10:00AM[0m [92mINF[0m CORS not configured, cross-origin browser requests will be refused [2mcomponent=[0mcontroller [2msubsystem=[0mcontroller.api [2mconfigure=[0mcontroller.api.security.cors.allow_origins
[2m10:00AM[0m [92mINF[0m metrics cache refresh started [2mcomponent=[0mcontroller [2msubsystem=[0mcontroller.health [2minterval=[0m15s
[2m10:00AM[0m [92mINF[0m embedded UI enabled [2mcomponent=[0mcontroller [2mpath=[0m/
[2m10:00AM[0m [92mINF[0m condition watcher started [2mcomponent=[0mcontroller [2msubsystem=[0mcontroller.notify
[2m10:00AM[0m [92mINF[0m metrics server started [2mcomponent=[0mcontroller [2msubsystem=[0mmetrics [2maddr=[0m0.0.0.0:9090
[2m10:00AM[0m [92mINF[0m http(s) server started [2mcomponent=[0mcontroller [2msubsystem=[0mmetrics [2maddress=[0m[::]:9090
[2m10:00AM[0m [92mINF[0m starting server [2mcomponent=[0mcontroller [2msubsystem=[0mcontroller.api
[2m10:00AM[0m [92mINF[0m http(s) server started [2mcomponent=[0mcontroller [2msubsystem=[0mcontroller.api [2maddress=[0m[::]:8080
[2m10:00AM[0m [92mINF[0m metrics cache initialized [2mcomponent=[0mcontroller [2msubsystem=[0mcontroller.health [2mnext_in=[0m15s
[2m10:00AM[0m [92mINF[0m registered in component registry [2mcomponent=[0mcontroller [2msubsystem=[0mcontroller.heartbeat [2mhostname=[0mweb-03 [2mtype=[0mcontroller [2mkey=[0mcontroller.web-03
```

With `--json`, the command prints:

```json
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"CORS not configured, cross-origin browser requests will be refused","component":"controller","subsystem":"controller.api","configure":"controller.api.security.cors.allow_origins"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"metrics cache refresh started","component":"controller","subsystem":"controller.health","interval":"15s"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"embedded UI enabled","component":"controller","path":"/"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"condition watcher started","component":"controller","subsystem":"controller.notify"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"metrics server started","component":"controller","subsystem":"metrics","addr":"0.0.0.0:9090"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"http(s) server started","component":"controller","subsystem":"metrics","address":"[::]:9090"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"starting server","component":"controller","subsystem":"controller.api"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"http(s) server started","component":"controller","subsystem":"controller.api","address":"[::]:8080"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"metrics cache initialized","component":"controller","subsystem":"controller.health","next_in":"15s"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"registered in component registry","component":"controller","subsystem":"controller.heartbeat","hostname":"web-03","type":"controller","key":"controller.web-03"}
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
