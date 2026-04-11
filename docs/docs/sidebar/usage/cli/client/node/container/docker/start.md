# Start

Start a stopped container on the target node:

```bash
$ osapi client node container docker start --id a1b2c3d4e5f6

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  MESSAGE
  server1   changed  true     container started

  1 host: 1 changed
```

Start a container by name:

```bash
$ osapi client node container docker start --id my-nginx
```

Target a specific host:

```bash
$ osapi client node container docker start --id my-nginx --target web-01
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node container docker start --id a1b2c3d4e5f6 --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--id`         | Container ID or name to start (**required**)             |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
