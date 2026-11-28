# Start

Start each component of osapi on this host.

This runs the embedded NATS server, the API server and the agent together. Use
it for a deployment on one host. For a deployment on more than one host, start
the NATS server, the controller and each agent alone.

The NATS server starts first, and waits until it is ready. Then osapi makes the
streams and buckets, starts the API server and starts the agent. On `SIGINT` or
`SIGTERM`, the three components stop at the same time, in 10 seconds or less.
Then osapi closes the NATS connections and the telemetry exporters.

```bash
osapi start
```

## Examples

```bash
osapi start
```

## Result

The command prints:

```ansi
$ osapi start
[2m10:00AM[0m [92mINF[0m nats server started successfully [2mcomponent=[0mnats
[2m10:00AM[0m [92mINF[0m embedded NATS server running [2mcomponent=[0mnats [2maddr=[0mlocalhost:4222 [2mstore_dir=[0m.nats/jetstream/ [2mnamespace=[0mosapi [2mauth.type=[0mnone
[2m10:00AM[0m [92mINF[0m metrics server started [2mcomponent=[0mnats [2msubsystem=[0mmetrics [2maddr=[0m0.0.0.0:9092
[2m10:00AM[0m [92mINF[0m http(s) server started [2mcomponent=[0mnats [2msubsystem=[0mmetrics [2maddress=[0m[::]:9092
[2m10:00AM[0m [92mINF[0m registered in component registry [2mcomponent=[0mnats [2msubsystem=[0mnats.heartbeat [2mhostname=[0mweb-03 [2mtype=[0mnats [2mkey=[0mnats.web-03
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
[2m10:00AM[0m [92mINF[0m starting node agent [2mcomponent=[0magent [2msubsystem=[0magent
[2m10:00AM[0m [92mINF[0m agent configuration [2mcomponent=[0magent [2msubsystem=[0magent [2mmachine_id=[0md530664edc1c431eb5ef1e9dfd6617a5 [2mhostname=[0mweb-03 [2mmax_jobs=[0m10 [2mlabels=[0mmap[group:web.dev.us-east]
[2m10:00AM[0m [92mINF[0m registered in agent registry [2mcomponent=[0magent [2msubsystem=[0magent.heartbeat [2mmachine_id=[0md530664edc1c431eb5ef1e9dfd6617a5 [2mhostname=[0mweb-03 [2mkey=[0magents.d530664edc1c431eb5ef1e9dfd6617a5 [2mttl=[0m30s
[2m10:00AM[0m [92mINF[0m node agent started successfully [2mcomponent=[0magent [2msubsystem=[0magent
[2m10:00AM[0m [92mINF[0m metrics server started [2mcomponent=[0magent [2msubsystem=[0mmetrics [2maddr=[0m0.0.0.0:9091
[2m10:00AM[0m [92mINF[0m http(s) server started [2mcomponent=[0magent [2msubsystem=[0mmetrics [2maddress=[0m[::]:9091
```

With `--json`, the command prints:

```json
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"nats server started successfully","component":"nats"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"embedded NATS server running","component":"nats","addr":"localhost:4222","store_dir":".nats/jetstream/","namespace":"osapi","auth.type":"none"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"metrics server started","component":"nats","subsystem":"metrics","addr":"0.0.0.0:9092"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"http(s) server started","component":"nats","subsystem":"metrics","address":"[::]:9092"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"registered in component registry","component":"nats","subsystem":"nats.heartbeat","hostname":"web-03","type":"nats","key":"nats.web-03"}
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
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"starting node agent","component":"agent","subsystem":"agent"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"agent configuration","component":"agent","subsystem":"agent","machine_id":"d530664edc1c431eb5ef1e9dfd6617a5","hostname":"web-03","max_jobs":10,"labels":{"group":"web.dev.us-east"}}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"registered in agent registry","component":"agent","subsystem":"agent.heartbeat","machine_id":"d530664edc1c431eb5ef1e9dfd6617a5","hostname":"web-03","key":"agents.d530664edc1c431eb5ef1e9dfd6617a5","ttl":"30s"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"node agent started successfully","component":"agent","subsystem":"agent"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"metrics server started","component":"agent","subsystem":"metrics","addr":"0.0.0.0:9091"}
{"time":"2026-01-15T10:00:00Z","level":"INFO","msg":"http(s) server started","component":"agent","subsystem":"metrics","address":"[::]:9091"}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](output.md) tells how the result is printed, and what `--json` gives.
