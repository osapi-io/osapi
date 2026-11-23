# Get

Show one method of an endpoint group: its route, fields, output, examples and
errors.

Each input field shows where the request carries it: the path, the query, a form
or the body. `--json` has the JSON Schema of the input and the output.
`osapi behavior method get` shows a method of a behavior the same way.

```bash
osapi controller endpoint method get [flags]
```

## Examples

```bash
osapi controller endpoint method get --name agent --method drain
osapi controller endpoint method get --name agent --method drain --json
```

## Result

The command prints:

```ansi
$ osapi controller endpoint method get --name agent --method drain
[1;38;2;0;213;255mMethod:[0m
  [38;2;10;143;176mName:[0m          drain
  [38;2;10;143;176mDescription:[0m   Drain an agent
  [38;2;10;143;176mHTTP:[0m          POST
  [38;2;10;143;176mRoute:[0m         /api/agent/{hostname}/drain
  [38;2;10;143;176mPermission:[0m    agent:write
  [38;2;10;143;176mAdmin only:[0m    yes
  [38;2;10;143;176mOperation ID:[0m  DrainAgent
  [38;2;10;143;176mCommand:[0m       osapi client agent drain
  [38;2;10;143;176mInput type:[0m    DrainInput
  [38;2;10;143;176mOutput type:[0m   none: it answers with no body
  [38;2;10;143;176mRefuses:[0m       a hostname over 255 characters

[1;38;2;255;0;255mInput (1):[0m
  [38;2;176;48;176mFIELD     [0m[38;2;176;48;176mIN    [0m[38;2;176;48;176mTYPE    [0m[38;2;176;48;176mREQUIRED  [0m[38;2;176;48;176mCONSTRAINTS                   [0m[38;2;176;48;176mDESCRIPTION[0m
  hostname  path  [1mstring[0m  yes       min length 1, max length 255  The hostname of the agent to drain.

[1;38;2;255;209;0mOutput (0):[0m
  [38;2;138;138;138mNone.[0m

[1;38;2;69;174;23mExamples (1):[0m
  [38;2;58;143;20mFLAGS[0m
  --hostname web-01

[1;38;2;107;92;255mErrors (2):[0m
  [38;2;89;76;216mCODE  [0m[38;2;89;76;216mREASON     [0m[38;2;89;76;216mDOC[0m
  404   not_found  Agent not found.
  409   conflict   Agent already in requested state.
```

With `--json`, the command prints:

```json
{
  "name": "drain",
  "doc": "Drain an agent",
  "http_method": "POST",
  "route": "/api/agent/{hostname}/drain",
  "permission": "agent:write",
  "admin_only": true,
  "operation_id": "DrainAgent",
  "command": "osapi client agent drain",
  "input_type": "DrainInput",
  "output_type": "",
  "fields": [
    {
      "field": "hostname",
      "in": "path",
      "type": "string",
      "required": true,
      "constraints": "min length 1, max length 255",
      "description": "The hostname of the agent to drain."
    }
  ],
  "results": [],
  "examples": [
    {
      "flags": "--hostname web-01",
      "doc": ""
    }
  ],
  "refuses": ["a hostname over 255 characters"],
  "errors": [
    {
      "code": 404,
      "reason": "not_found",
      "doc": "Agent not found."
    },
    {
      "code": 409,
      "reason": "conflict",
      "doc": "Agent already in requested state."
    }
  ],
  "input": null,
  "output": null
}
```

## Flags

| Flag       | Description                                               | Default |
| ---------- | --------------------------------------------------------- | ------- |
| `--method` | The method, from osapi controller endpoint get (required) |         |
| `--name`   | The endpoint group, such as agent (required)              |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
