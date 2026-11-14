# Health

```sh
osapi client health status --json      # every component, and the controller's view of NATS
osapi client health ready --json       # is the controller ready to serve
osapi client agent list --json         # agents that have checked in, with conditions
osapi client agent get --hostname <h> --json
osapi client nats status --json
osapi client job get --id <job> --json # one job, and each host's result
```

A host missing from `agent list` has not sent a heartbeat within
`nats.registry.ttl`: its agent is down, cannot reach NATS, or uses a different
`namespace`. Compare that agent's `config list` with the server's.

An agent reports three conditions, high load, memory pressure and disk
pressure, with the numbers that tripped them. The thresholds are under
`agent.conditions`.

A drained agent takes no new jobs: `client agent undrain` returns it.

A job whose hosts all report `timeout` reached no agent in
`controller.api.job_timeout`. Check `agent list` and the target.

Start any component with `--debug` for its log, and look there before
changing config.
