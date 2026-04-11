# Remove

Remove a container from the target node:

```bash
$ osapi client node container docker remove --id a1b2c3d4e5f6

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  MESSAGE
  server1   changed  true     container removed

  1 host: 1 changed
```

Force removal of a running container:

```bash
$ osapi client node container docker remove --id my-nginx --force
```

Target a specific host:

```bash
$ osapi client node container docker remove --id my-nginx --target web-01
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node container docker remove --id a1b2c3d4e5f6 --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--id`         | Container ID or name to remove (**required**)            |         |
| `--force`      | Force removal of a running container                     |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
