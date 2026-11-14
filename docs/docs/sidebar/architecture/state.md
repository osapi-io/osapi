---
sidebar_position: 5
---

# State and storage

osapi keeps its state in NATS JetStream and nowhere else: there is no database.
The controller and the agents are stateless apart from what they write there,
and the files an agent manages on its own host. Losing a controller loses
nothing; losing JetStream loses everything below.

The NATS setup creates each store at startup if it does not exist. Every name is
set in the `nats` section of the config, and the ones below are the names in the
shipped configuration; each store's retention and its storage, memory or file,
are set there too.

| Store              | Kind         | Holds                                                         | Kept for                 | Losing it                                                    |
| ------------------ | ------------ | ------------------------------------------------------------- | ------------------------ | ------------------------------------------------------------ |
| `JOBS`             | Stream       | A notification per job, on `jobs.query.*` and `jobs.modify.*` | `nats.stream.max_age`    | Jobs not yet picked up are never run                         |
| `JOBS-DLQ`         | Stream       | Delivery advisories of jobs that ran out of retries           | `nats.dlq.max_age`       | The record of which jobs gave up                             |
| `job-queue`        | KV           | Each job's definition and its status events                   | `nats.kv.ttl`            | Job history; `job get` and `job retry` stop working for them |
| `job-responses`    | KV           | Each agent's result for each job                              | `nats.kv.ttl`            | Results of jobs still being read                             |
| `agent-registry`   | KV           | Heartbeats from every agent, controller and NATS server       | `nats.registry.ttl`, 30s | Nothing lasting; components rewrite it every 10 seconds      |
| `agent-facts`      | KV           | Each agent's facts: interfaces, routes and the like           | `nats.facts.ttl`         | Nothing lasting; agents collect again                        |
| `agent-state`      | KV           | Operator actions: drain flags and timeline events             | No expiry                | Drains are lifted and the timeline is gone                   |
| `file-objects`     | Object store | Files uploaded with `osapi client file upload`                | Until deleted            | Uploaded files; deploys of them fail                         |
| `file-state`       | KV           | What osapi deployed to each host, with each file's checksum   | No expiry                | osapi stops knowing which files on a host it manages         |
| `AUDIT`            | Stream       | One entry per authenticated API request                       | `nats.audit.max_age`     | The audit trail                                              |
| `agent-enrollment` | KV           | Agents waiting to be accepted, and accepted agents' keys      | No expiry                | Every agent must enroll and be accepted again                |

The stores an operation needs and the config does not name are not created, and
the operations that use them are turned off: without `file-state`, for instance,
the agent turns its file operations off and the sysctl behavior reports
`skipped`. `agent-enrollment` is created by the controller, and only when PKI is
on.

## What lives on the hosts

An agent writes the files it manages, such as `osapi-<key>.conf` under
`/etc/sysctl.d`, and records each in `file-state` with its checksum. That record
is how osapi tells its own files from ones it did not write: it lists, updates
and removes only what it recorded, and leaves everything else alone. An agent
also keeps its PKI keypair in `agent.pki.key_dir`.

## Namespaces

Several osapi deployments can share one NATS server. Each sets a namespace in
`nats.server.namespace`, `controller.nats.namespace` and `agent.nats.namespace`,
which prefixes every subject and every store name, so their jobs, registries and
audit logs never mix.

## Sizing

Each agent creates JetStream consumers for its machine ID, `_any`, `_all` and
every prefix of each label, for both reads and changes. One label three levels
deep is six consumers; five labels like it and the base subjects come to about
thirty-six per agent. A fleet in the hundreds, or one that needs the stores to
survive a host, belongs on an external NATS cluster with file storage and
replicas set in each store's section. [Deployment](deployment.md) has the
layouts.

## Design details

- [The job system](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md),
  for how jobs, status events and responses are keyed.
- [Configuration](https://github.com/osapi-io/specs/blob/main/components/osapi/configuration.md),
  for every store setting, and `osapi config list` for their values on a host.
