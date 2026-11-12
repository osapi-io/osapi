# Behaviors

A behavior is an area of a host osapi manages, such as `sysctl`. Each has
methods, and each method has a typed input and output.

```sh
osapi behavior list --json                     # every behavior and its methods
osapi behavior describe sysctl --json          # each method's route, permission and command
osapi behavior describe sysctl update --json   # one method's input and output schema
```

`describe --json` gives, per method:

- `command`: the CLI command to run, such as `osapi client node sysctl update`.
- `input`: JSON Schema. Each property is a flag of the same name; `required`
  lists the ones you must pass. `pattern`, `enum`, `minLength` and the rest are
  checked before the job is queued, so a value that breaks them fails with
  `invalid_request` and nothing runs.
- `output`: JSON Schema for one result row. Every row also carries `hostname`,
  `status` and, when it failed, `error`. A write's row carries `changed`, which
  is `false` when the host was already in that state.
- `permission`: what your token needs, such as `sysctl:write`.

A read (`verb: read`) changes nothing. A write is safe to repeat: creating
something that exists, or deleting something absent, succeeds with
`changed: false`. Updating something absent fails.

Controller commands that are not behaviors, `client job`, `client audit`,
`client agent` and `client health`, are found with `osapi help client --json`.
