# Inspect

Get detailed information about a specific container:

```bash
$ osapi client node container docker inspect --id a1b2c3d4e5f6

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  ID            NAME      IMAGE         STATE    CREATED               HEALTH   PORTS             MOUNTS             NETWORK
  a1b2c3d4e5f6  my-nginx  nginx:latest  running  2024-01-15T10:30:00Z           8080:80,8443:443  /data:/var/lib/data  172.17.0.2
```

Inspect a container by name:

```bash
$ osapi client node container docker inspect --id my-nginx
```

Target a specific host:

```bash
$ osapi client node container docker inspect --id my-nginx --target web-01
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node container docker inspect --id a1b2c3d4e5f6 --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--id`         | Container ID or name to inspect (**required**)           |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
