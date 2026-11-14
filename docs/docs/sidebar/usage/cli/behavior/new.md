# New

Write a new behavior's standard files into an osapi checkout and register it.
The result builds, validates and passes every test before you change anything;
then change the types, the provider and the examples to what the behavior
manages. Writes nothing when the name is taken.

```bash
$ osapi behavior new --name motd --description "The message shown at login"
```

A behavior that belongs to a group, such as `network`, is nested under it: its
routes are `/api/node/{hostname}/network/<name>`, its commands
`osapi client node network <name>`, and its directory
`internal/behaviors/network/<name>/`.

```bash
$ osapi behavior new --group network --name neighbors --description "The host's ARP neighbors"
```

[Your first behavior](../../../development/your-first-behavior.md) walks through
the rest.

## Flags

| Flag            | Description                                                  |
| --------------- | ------------------------------------------------------------ |
| `--name`        | The behavior's name: lowercase letters and digits (required) |
| `--description` | What it manages, in one line (required)                      |
| `--root`        | The osapi checkout to write into (default `.`)               |
