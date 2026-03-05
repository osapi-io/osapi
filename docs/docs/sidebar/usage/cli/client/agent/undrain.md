# Undrain

Resume accepting jobs on a drained or cordoned agent:

```bash
$ osapi client agent undrain --hostname web-01

  Hostname: web-01
  Status: Ready
  Message: Agent undrain initiated
```

The agent re-subscribes to NATS JetStream consumers and transitions back to
`Ready`.

## Flags

| Flag         | Description                      | Required |
| ------------ | -------------------------------- | -------- |
| `--hostname` | Hostname of the agent to undrain | Yes      |
