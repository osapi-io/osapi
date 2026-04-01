# List

List SSH authorized keys for a user:

```bash
$ osapi client node user ssh-key list --target web-01 --name deploy

  HOSTNAME  TYPE         FINGERPRINT              COMMENT      STATUS
  web-01    ssh-ed25519  SHA256:abc123...          user@laptop  ok
  web-01    ssh-rsa      SHA256:def456...          deploy-ci    ok
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node user ssh-key list --target web-01 --name deploy --json
{"results":[{"hostname":"web-01","keys":[{"type":"ssh-ed25519","fingerprint":"SHA256:abc123...","comment":"user@laptop"}],"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Username to list SSH keys for (required)                 |         |
| `-j, --json`   | Output raw JSON response                                 |         |
