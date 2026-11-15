# Targets

Every `client node` command takes `--target` (`-T`):

| Value           | Means                             |
| --------------- | --------------------------------- |
| `web-01`        | that host                         |
| `_any`          | one host, whichever answers first |
| `_all`          | every host (the default)          |
| `group:web.dev` | every host whose label matches    |

A command against more than one host returns one result per host. Hosts that do
not answer in time come back as `timeout`, and the command still succeeds. Check
each row's `status`, not the exit code alone.

Start with one host. Run the change against `_all` only after it has done what
you expected on one.
