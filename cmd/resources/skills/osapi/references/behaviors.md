# Behaviors

A behavior is an area of a host osapi manages, such as `sysctl`. Each has
methods, and each method has a typed input and output.

```sh
osapi behavior list --json                              # every behavior and its methods
osapi behavior get --name sysctl --json                 # surfaces, methods, chains, telemetry
osapi behavior get --name sysctl --method update --json # one method's input and output schema
```

`behavior get --json` gives, for the behavior:

- `api` and `cli`: whether it is served over REST and has commands.
- `platforms`: where it runs. On any other platform its methods report
  `skipped`, which is not a failure.
- `sudo`: the programs it runs with sudo on the host.
- `uses` and `used_by`: the behaviors it is built on, and those built on it.
  Calling one that uses others runs those too.
- `telemetry`: the span and metric names its calls record, and which inputs are
  recorded on the span.

And per method:

- `command`: the CLI command to run, such as `osapi client next node sysctl update`.
  Empty when the behavior has no CLI; use `route` and the API instead.
- `route` and `http_method`: where the API serves it. Empty when it has no API.
- `input`: JSON Schema. Each property is a flag of the same name; `required`
  lists the ones you must pass. `pattern`, `enum`, `minLength` and the rest are
  checked before the job is queued, so a value that breaks them fails with
  `invalid_request` and nothing runs.
- `input_type` and `output_type`: the Go types the schemas come from.
- `output`: JSON Schema for one result row. Every row also carries `hostname`,
  `status` and, when it failed, `error`. A write's row carries `changed`, which
  is `false` when the host was already in that state.
- `permission`: what your token needs, such as `sysctl:write`.
- `examples`: sample inputs with the output one host returns for each.

A read (`verb: read`) changes nothing. A write is safe to repeat: creating
something that exists, or deleting something absent, succeeds with
`changed: false`. Updating something absent fails.

Controller commands that are not behaviors, `client job`, `client audit`,
`client agent` and `client health`, are found with `osapi help client --json`.
