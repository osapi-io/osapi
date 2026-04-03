# Stale

List deployments where the source object has been updated since the file was
last deployed. Shows which files need redeployment. Requires `file:read`
permission.

```bash
$ osapi client file stale

  Total: 2

  Stale Deployments:
  OBJECT       HOSTNAME  PATH                                           DEPLOYED              DEPLOYED SHA   CURRENT SHA
  hello-echo   web-01    /etc/systemd/system/osapi-hello-echo.service   2026-04-01T18:00:00Z  abc123def456…  789abc012def…
  my-ca-cert   web-02    /usr/local/share/ca-certificates/osapi-my-ca   2026-03-31T12:00:00Z  111222333444…  555666777888…
```

When all deployments are in sync:

```bash
$ osapi client file stale

  Total: 0

  All deployments are in sync.
```

## JSON Output

Use `--json` for raw JSON output:

```bash
$ osapi client file stale --json
{"stale":[...],"total":2}
```

## Flags

| Flag         | Description              | Default |
| ------------ | ------------------------ | ------- |
| `-j, --json` | Output raw JSON response |         |
