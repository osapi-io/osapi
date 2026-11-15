# Drain

Drain an agent so that it does not accept new jobs. The jobs that it runs now
continue until they are complete:

```bash
$ osapi client agent drain --hostname web-01

  Hostname: web-01
  Status: Draining
  Message: Agent drain initiated
```

The agent changes from `Ready` to `Draining`. When all its running jobs are
complete, the state is `Cordoned`. The agent continues to run and to send
heartbeats. It only stops the intake of new work from the job queue.

Use `agent undrain` to accept jobs again.

## Flags

| Flag         | Description                    | Required |
| ------------ | ------------------------------ | -------- |
| `--hostname` | Hostname of the agent to drain | Yes      |
