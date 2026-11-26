# Health

```sh
osapi client health status --json      # every component, with its sub-components, and the controller's checks
osapi client health ready --json       # is the controller ready to serve
osapi client agent list --json         # agents that have checked in, with conditions
osapi client agent get --hostname <h> --json
osapi client health status --component nats --json        # the NATS server, with its sub-components
osapi client job get --id <job> --json # one job, and each host's result
```

A host missing from `agent list` has not sent a heartbeat within
`nats.registry.ttl`: its agent is down, cannot reach NATS, or uses a different
`namespace`. Compare that agent's `config list` with the server's.

An agent reports three conditions, high load, memory pressure and disk pressure,
with the numbers that tripped them. The thresholds are under `agent.conditions`.

A drained agent takes no new jobs: `client agent undrain` returns it.

A job whose hosts all report `timeout` reached no agent in
`controller.api.job_timeout`. Check `agent list` and the target.

`health status` counts the dead letters: jobs whose delivery to an agent used up
its retries. List them, newest first, and send one back to the queue:

```sh
osapi client job deadletter list --json          # each with its job, target and error
osapi client job deadletter list --limit 5 --json
osapi client job retry --id <job> --json         # a new job with the same operation
```

A dead letter stays in the stream until its max age (`nats.dlq.max_age`); retry
makes a new job and leaves the dead letter as it was.

Start any component with `--debug` for its log, and read that log before you
change the config.
