# Components

| Component   | Start one                  | Listens on, by default                 |
| ----------- | -------------------------- | -------------------------------------- |
| NATS server | `osapi nats server start`  | 4222 (`nats.server.port`)              |
| Controller  | `osapi controller start`   | 8080 (`controller.api.port`)           |
| Agent       | `osapi agent start`        | no API port                            |

`osapi start` runs all three in one process, which is the usual way on a single
host. Run them separately to put the NATS server and controller on one host and
an agent on every other.

Every component reads the same config file, `-f` or `--osapi-file`, default
`/etc/osapi/osapi.yaml`. The controller and each agent connect to NATS through
`controller.nats.*` and `agent.nats.*`, which must point at the server and use
the same `namespace` as `nats.server.namespace`.

Start order, when separate: the NATS server, then the controller, then agents.

An agent's name is `agent.hostname`, or the host's own name when that is empty.
Its `agent.labels`, at most five, are what a `group:web` style target matches.

Each component also has a metrics port, `*.metrics.port`, which serves
`/metrics`, `/health` and `/health/ready` without a token, for a process
supervisor or a load balancer.