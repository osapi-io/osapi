# Running the fleet

## The agents

```sh
osapi client agent list --json                  # each agent: state, labels, conditions, config_behind
osapi client agent get --hostname web-01 --json # one agent, with its timeline
osapi client agent drain --hostname web-01      # it takes no new job; undrain to resume
osapi client agent config update --hostname web-01 --labels group=db --json
osapi client agent facts list --json            # the keys of @fact references
osapi agent collector list --json               # what the agent collects, by category
osapi agent collector get --name <name> --json
osapi agent collector spec --json               # the parts of a collector
```

## The audit log

Each change is an entry: a request of the API, or a change that did not go
through it, with an actor (`system`, `join-token`, `agent`, `local`).

```sh
osapi client audit list --writes --json         # changes only; --user, --status, --operation filter
osapi client audit get --id <id> --json
osapi client audit export --json                # each entry, with no pages
```

## Files and secrets

```sh
osapi client file upload --name nginx.conf --file ./nginx.conf --content-type template --json
osapi client file list --json
osapi client file stale list --json             # deployments whose file changed since
osapi client secret vault list --json
osapi client secret create --vault default --key registry-token --value - --json   # the value on standard input
osapi client secret update --vault default --key registry-token --value - --json   # a new value for a key that exists
osapi client secret list --vault default --json # names and metadata, never a value
```

`osapi client secret reveal` prints a value: it needs `secret:reveal`, which
only admin has. Never print a secret in an answer.

## Traces and the catalog

```sh
osapi client telemetry trace get --id <trace id> --json   # the span tree of a request; a command prints its trace ID
osapi client catalog get --json                            # what this controller can do
```
