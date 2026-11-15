# Undrain

Let a drained or cordoned agent accept jobs again:

```bash
$ osapi client agent undrain --hostname web-01

  Hostname: web-01
  Status: Ready
  Message: Agent undrain initiated
```

The agent subscribes to the NATS JetStream consumers again and changes back to
`Ready`.

## Flags

| Flag         | Description                      | Required |
| ------------ | -------------------------------- | -------- |
| `--hostname` | Hostname of the agent to undrain | Yes      |
