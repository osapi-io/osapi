# The controller's own operations

Some operations act on osapi itself, not on a host: the agents, the jobs, the
audit log, the uploaded files, health, facts and traces. They run on the
controller, in the request. They have no target and no result for each host.

The binary describes them as endpoint groups:

```sh
osapi controller endpoint list --json                 # the groups the binary defines, and their methods
osapi controller endpoint get --name agent --json     # each method's verb, path, permission and operationId
osapi controller endpoint get --name agent --openapi  # the group's part of the OpenAPI spec
```

Only some groups are defined yet. `controller endpoint list` shows which. For a group that
is not in the list, ask the CLI: `osapi help client <group> --json`.

The commands for them are under `osapi client`, such as `osapi client agent
list`, `osapi client agent pending`, `osapi client agent drain --hostname web-01`
and `osapi client job get --id <id>`. Pass `--json`. A command of a group that
`endpoint list` shows has one flag for each input field of its method, named as
the field is in JSON, with `_` as `-`.

The permission of each method is in `controller endpoint get --json`. A token without it
gets 403. Reads need `<group>:read` and changes need `<group>:write`, unless the
method names a different one.

A 409 says that the thing is not in a state that allows the operation, such as
an agent that is already drained. A 500 is a failure of the controller. Report
it. Do not repeat a change to fix it.
