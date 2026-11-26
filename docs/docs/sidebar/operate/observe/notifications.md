# Notifications

OSAPI monitors component health through condition evaluation and notifies when
conditions change. The notification system watches the registry KV bucket for
condition transitions and dispatches events through a pluggable backend.

## How it works

Every component (agent, controller, NATS server) evaluates conditions on each
heartbeat and writes them to the registry KV bucket. A watcher on the controller
detects transitions:

- **Fired**: a condition becomes active, for example when `DiskPressure` goes
  over the limit
- **Resolved**: a condition becomes inactive, for example when the disk use goes
  below the limit
- **Unreachable**: the heartbeat of a component expires (the TTL ends)

Active conditions are re-fired at a configurable interval so they remain visible
in logs and alerts.

## Conditions

| Condition               | Components       | Description                                               |
| ----------------------- | ---------------- | --------------------------------------------------------- |
| `MemoryPressure`        | agent            | Host memory usage exceeds threshold (default 90%)         |
| `HighLoad`              | agent            | Load average exceeds CPU count × multiplier (default 2.0) |
| `DiskPressure`          | agent            | Any disk usage exceeds threshold (default 90%)            |
| `ProcessMemoryPressure` | agent, api, nats | Process RSS exceeds configured byte threshold             |
| `ProcessHighCPU`        | agent, api, nats | Process CPU usage exceeds configured percent threshold    |
| `ComponentUnreachable`  | agent, api, nats | Heartbeat expired (TTL timeout)                           |

Only agents calculate the host conditions. All components calculate the process
conditions. The watcher sends `ComponentUnreachable` when the TTL of a heartbeat
expires. It is not in the registration of the component, because the component
is already gone.

## Notifier backends

The notification system uses a pluggable `Notifier` interface. Currently one
backend is available:

### Log (default)

Writes condition events to the structured log. Fired conditions log at WARN
level, resolved conditions at INFO. `component` is the controller, which writes
the line. `source` is the component that has the condition:

```
WRN condition fired component=controller source=agent hostname=web-01 condition=DiskPressure active=true reason="/ 92% used"
INF condition resolved component=controller source=agent hostname=web-01 condition=DiskPressure active=false
WRN condition fired component=controller source=nats hostname=nats-01 condition=ComponentUnreachable active=true reason="heartbeat expired"
```

### Future backends

- **Slack**: post to a webhook URL
- **Email**: send through SMTP
- **Webhook**: POST to a URL that you set

## Re-notification

By default, a condition fires once when it becomes active and once when it
resolves. To keep active conditions visible, configure `renotify_interval`:

```yaml
controller:
  notifications:
    enabled: true
    notifier: 'log'
    renotify_interval: '5m'
```

With `renotify_interval: '5m'`, an active `DiskPressure` condition re-fires
every 5 minutes until resolved. Uses Go duration format (`1m`, `5m`, `1h`). Set
to `'0'` to disable re-notification.

## Configuration

| Key                 | Env Variable                                       | Description                           |
| ------------------- | -------------------------------------------------- | ------------------------------------- |
| `enabled`           | `OSAPI_CONTROLLER_NOTIFICATIONS_ENABLED`           | Enable the watcher (default: `false`) |
| `notifier`          | `OSAPI_CONTROLLER_NOTIFICATIONS_NOTIFIER`          | Backend: `"log"` (default)            |
| `renotify_interval` | `OSAPI_CONTROLLER_NOTIFICATIONS_RENOTIFY_INTERVAL` | Re-fire interval (default: `"0"`)     |

Refer to [Configuration](../plan.md#the-config-file) for the full reference.

## Architecture

The watcher is a background goroutine in the controller. It watches the registry
KV bucket with NATS KV Watch. At each update, it compares the previous
conditions with the current conditions, and sends a `ConditionEvent` for each
change.

The watcher can move to a separate process. It needs only access to NATS KV and
a `Notifier` implementation.
