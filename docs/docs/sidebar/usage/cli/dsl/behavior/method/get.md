# Get

Show one method of a behavior: its fields, constraints, output and examples.

Each input field is a flag of its command and a field of its request. `--json`
has the JSON Schema of the input and the output, with each constraint that osapi
checks before it queues the job.

```bash
osapi dsl behavior method get [flags]
```

## Examples

```bash
osapi dsl behavior method get --name sysctl --method update
osapi dsl behavior method get --name sysctl --method update --json
```

## Result

The command prints:

```ansi
$ osapi dsl behavior method get --name sysctl --method update
[1;38;2;0;213;255mMethod:[0m
  [38;2;10;143;176mName:[0m         update
  [38;2;10;143;176mVerb:[0m         write
  [38;2;10;143;176mDescription:[0m  Set a kernel parameter, manage it, and apply it
  [38;2;10;143;176mHTTP:[0m         PUT
  [38;2;10;143;176mRoute:[0m        /api/node/{hostname}/sysctl/{key}
  [38;2;10;143;176mPermission:[0m   sysctl:write
  [38;2;10;143;176mCommand:[0m      osapi client node sysctl update
  [38;2;10;143;176mInput type:[0m   UpdateInput
  [38;2;10;143;176mOutput type:[0m  MutationResult

[1;38;2;255;0;255mInput (2):[0m
  [38;2;176;48;176mFIELD  [0m[38;2;176;48;176mTYPE    [0m[38;2;176;48;176mREQUIRED  [0m[38;2;176;48;176mCONSTRAINTS                                         [0m[38;2;176;48;176mDESCRIPTION[0m
  key    [1mstring[0m  yes       pattern ^[a-zA-Z0-9][a-zA-Z0-9._-]*$, min length 1  Parameter name, such as net.ipv4.ip_forward
  value  [1mstring[0m  yes       pattern ^[^\r\n]*$                                  Value to set

[1;38;2;255;209;0mOutput (2):[0m
  [38;2;176;143;0mFIELD    [0m[38;2;176;143;0mTYPE     [0m[38;2;176;143;0mDESCRIPTION[0m
  key      [1mstring[0m   Parameter name
  changed  [1mboolean[0m  Whether the host changed

[1;38;2;69;174;23mExamples (1):[0m
  [38;2;58;143;20mEXAMPLE           [0m[38;2;58;143;20mINPUT                        [0m[38;2;58;143;20mOUTPUT[0m
  Lower swappiness  key=vm.swappiness, value=10  key=vm.swappiness, changed=true
```

With `--json`, the command prints:

```json
{
  "name": "update",
  "verb": "write",
  "doc": "Set a kernel parameter, manage it, and apply it",
  "http_method": "PUT",
  "route": "/api/node/{hostname}/sysctl/{key}",
  "permission": "sysctl:write",
  "command": "osapi client node sysctl update",
  "input_type": "UpdateInput",
  "output_type": "MutationResult",
  "fields": [
    {
      "field": "key",
      "type": "string",
      "required": true,
      "constraints": "pattern ^[a-zA-Z0-9][a-zA-Z0-9._-]*$, min length 1",
      "description": "Parameter name, such as net.ipv4.ip_forward"
    },
    {
      "field": "value",
      "type": "string",
      "required": true,
      "constraints": "pattern ^[^\\r\\n]*$",
      "description": "Value to set"
    }
  ],
  "results": [
    {
      "field": "key",
      "type": "string",
      "description": "Parameter name"
    },
    {
      "field": "changed",
      "type": "boolean",
      "description": "Whether the host changed"
    }
  ],
  "example_rows": [
    {
      "example": "Lower swappiness",
      "input": "key=vm.swappiness, value=10",
      "output": "key=vm.swappiness, changed=true"
    }
  ],
  "input": null,
  "output": null
}
```

## Flags

| Flag       | Description                                           | Default |
| ---------- | ----------------------------------------------------- | ------- |
| `--method` | The method, from osapi dsl behavior get (required)    |         |
| `--name`   | The behavior, from osapi dsl behavior list (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
