# List

List SSH authorized keys for a user:

```bash
$ osapi client node user ssh-key list --target web-01 --name deploy

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  TYPE         FINGERPRINT      COMMENT
  web-01    ok      ssh-ed25519  SHA256:abc123...  user@laptop
  web-01    ok      ssh-rsa      SHA256:def456...  deploy-ci

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node user ssh-key list --target _all --name deploy

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  TYPE         FINGERPRINT      COMMENT
  web-01    ok      ssh-ed25519  SHA256:abc123...  user@laptop
  web-02    ok      ssh-ed25519  SHA256:abc123...  user@laptop
  web-02    ok      ssh-rsa      SHA256:ghi789...  deploy-prod

  2 hosts: 2 ok
```

Hosts with no authorized keys are omitted from the output. Skipped hosts (e.g.,
unsupported platforms) show with their error.

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node user ssh-key list --target web-01 --name deploy --json
{"results":[{"hostname":"web-01","status":"ok","keys":[
{"type":"ssh-ed25519","fingerprint":"SHA256:abc123...",
"comment":"user@laptop"}]}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Username to list SSH keys for (required)                 |         |
| `-j, --json`   | Output raw JSON response                                 |         |
