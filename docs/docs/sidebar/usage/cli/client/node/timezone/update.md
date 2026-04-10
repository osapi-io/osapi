# update

Update the system timezone on a target node.

## Usage

```bash
osapi client node timezone update [flags]
```

## Flags

| Flag         | Type   | Default | Description                                |
| ------------ | ------ | ------- | ------------------------------------------ |
| `--target`   | string | `_any`  | Target hostname, `_all`, or label selector |
| `--timezone` | string |         | IANA timezone name (required)              |
| `--json`     | bool   | `false` | Output raw JSON response                   |

## Examples

```bash
# Set timezone on a specific host
osapi client node timezone update --target web-01 \
  --timezone America/New_York

# Set timezone on all hosts
osapi client node timezone update --target _all \
  --timezone UTC

# Get raw JSON output
osapi client node timezone update --target web-01 \
  --timezone UTC --json
```

## Output

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR
  web-01    ok      true
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
      "changed": true
    }
  ]
}
```
