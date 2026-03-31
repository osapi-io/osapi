# get

Get the current system timezone from a target node.

## Usage

```bash
osapi client node timezone get [flags]
```

## Flags

| Flag       | Type   | Default | Description                                |
| ---------- | ------ | ------- | ------------------------------------------ |
| `--target` | string | `_any`  | Target hostname, `_all`, or label selector |
| `--json`   | bool   | `false` | Output raw JSON response                   |

## Examples

```bash
# Get timezone from any available agent
osapi client node timezone get

# Get timezone from a specific host
osapi client node timezone get --target web-01

# Get timezone from all hosts
osapi client node timezone get --target _all

# Get raw JSON output
osapi client node timezone get --json
```

## Output

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  TIMEZONE          UTC_OFFSET
  web-01    America/New_York  -05:00
```

## JSON Output

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "hostname": "web-01",
      "status": "ok",
      "timezone": "America/New_York",
      "utc_offset": "-05:00"
    }
  ]
}
```
