# Image Remove

Remove a container image from the target node:

```bash
$ osapi client node container docker image-remove --image nginx:latest

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR  MESSAGE
  server1   ok      true            Image removed successfully
```

Force remove an image that may be in use:

```bash
$ osapi client node container docker image-remove \
    --image nginx:latest --force
```

Target a specific host:

```bash
$ osapi client node container docker image-remove \
    --image redis:7 --target web-01
```

Remove on all hosts:

```bash
$ osapi client node container docker image-remove \
    --image nginx:latest --target _all
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node container docker image-remove --image nginx:latest --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--image`      | Image name or ID to remove (**required**)                |         |
| `--force`      | Force removal even if image is in use                    | `false` |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
