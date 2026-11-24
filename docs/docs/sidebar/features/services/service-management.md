# Service management

osapi reads the systemd services of each host. It deploys, removes, starts,
stops, restarts, enables and disables the services that it deploys itself. It is
the [service](../../reference/behaviors/services/service/service.md) behavior.

A unit file goes to the fleet in two steps. Upload the file one time with
`osapi client file upload`. Then deploy it by its name to each host. The agent
gets the object from the Object Store with the
[object](../../reference/behaviors/internal/object/object.md) building block,
and writes it with
[managedfile](../../reference/behaviors/internal/managedfile/managedfile.md).
The [chain](../../reference/chains/services/service.md) of the behavior shows
these calls.

## How it works

### List and get

gohai reads the services with `systemctl list-units` and
`systemctl list-unit-files` on each Linux host. Each service has its name, its
state (such as `running`, `exited` or `dead`), whether it starts at boot and its
description. Get reads one service from the same list, by its name with
`.service` or without it. A service that systemd has not loaded, such as one
that is stopped and disabled, is not in the list and is not found
(osapi-io/gohai#321). A Mac has no systemd, and reports skipped.

### The services that osapi deploys

osapi writes and acts on its own services only. The unit file of each is
`/etc/systemd/system/osapi-<name>.service`, and osapi keeps a record of each
file that it writes. Thus `--name app` is `osapi-app.service`, and osapi cannot
change a service that a package installed, such as `ssh`. A unit file with the
prefix that osapi did not write, such as the agent's own `osapi-agent.service`,
is not osapi's either: create refuses it. An update, a delete or an action on a
service that osapi did not deploy is not found.

- **Create** writes the object as the unit file, and runs
  `systemctl daemon-reload` when the file changed. A unit file that has the
  content reports `changed: false`. The service does not start: start and enable
  are their own actions.
- **Update** writes a different object over the unit file of a service that
  osapi deployed.
- **Delete** runs `systemctl disable --now`, removes the unit file and reloads
  systemd. A service with no unit file reports `changed: false`.

When `systemctl daemon-reload` fails, a retry with the same input reloads
systemd again. After a create or an update, the unit file stays, and its record
has a mark that tells that the reload did not finish. After a delete, the unit
file comes back with the mark. The next write of the service reports a change
and removes the mark.

### Actions

Each action runs `systemctl <action> osapi-<name>.service` on the Debian family.

| Action  | Reports `changed: false` when        |
| ------- | ------------------------------------ |
| start   | the service runs                     |
| stop    | the service does not run             |
| enable  | the service starts at boot           |
| disable | the service does not start at boot   |
| restart | never: each restart changes the host |

## CLI usage

```bash
# Upload the unit file one time
osapi client file upload --name app.service --file ./app.service

# List the services of a host
osapi client node service list --target web-01

# Get one service
osapi client node service get --target web-01 --name ssh

# Deploy the unit file as osapi-app.service
osapi client node service create --target web-01 --name app --object app.service

# Start it and make it start at boot
osapi client node service start --target web-01 --name app
osapi client node service enable --target web-01 --name app

# Deploy a new version of the unit file
osapi client node service update --target web-01 --name app --object app-v2.service

# Restart it on each host of a group
osapi client node service restart --target group:web --name app

# Stop, disable and remove it
osapi client node service delete --target web-01 --name app
```

All commands accept `--json` to give the raw JSON output. All commands accept
broadcast targets, such as `_all` or a label selector, and give one result for
each host.

## Permissions and platforms

The generated page of each behavior gives the permission, the platforms and the
route of each method:
[service](../../reference/behaviors/services/service/service.md),
[object](../../reference/behaviors/internal/object/object.md). The pages come
from the definitions, so they stay true when a method changes, and this page
does not copy them. [Authentication](../platform/authentication.md) tells which
permissions each role has.

## Related

- [service](../../reference/behaviors/services/service/service.md): each method,
  with its input, its output and its examples
- [CLI Reference](../../usage/cli/client/node/service/service.mdx): the service
  commands
- [SDK Reference](../../sdk/services/service.md): the service of the SDK
- [File management](../config/file-management.md): upload the unit file
