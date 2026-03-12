# Stop

Stop a running container on the target node:

```bash
$ osapi client container stop --id a1b2c3d4e5f6

  Job ID:   550e8400-e29b-41d4-a716-446655440000

  Hostname: server1
  Message:  container stopped
```

Stop with a custom timeout (seconds to wait before killing):

```bash
$ osapi client container stop --id my-nginx --timeout 30
```

Target a specific host:

```bash
$ osapi client container stop --id my-nginx --target web-01
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client container stop --id a1b2c3d4e5f6 --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--id`         | Container ID or name to stop (**required**)              |         |
| `--timeout`    | Seconds to wait before killing the container             | `10`    |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
