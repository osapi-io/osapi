# Drain

Drain an agent to stop it from accepting new jobs. In-flight jobs continue to
completion:

```bash
$ osapi client agent drain --hostname web-01

  Hostname: web-01
  Status: Draining
  Message: Agent drain initiated
```

The agent transitions from `Ready` to `Draining`. Once all in-flight jobs
finish, the state becomes `Cordoned`. The agent stays running and continues
sending heartbeats -- it just stops pulling new work from the job queue.

Use `agent undrain` to resume accepting jobs.

## Flags

| Flag         | Description                    | Required |
| ------------ | ------------------------------ | -------- |
| `--hostname` | Hostname of the agent to drain | Yes      |
