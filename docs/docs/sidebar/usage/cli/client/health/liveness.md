# Liveness

Check if the API server process is running:

```bash
$ osapi client health liveness

  Status: ok
```

Use `--json` for raw JSON output:

```bash
$ osapi client health liveness --json
{"status":"ok"}
```
