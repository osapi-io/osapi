# Get

Describe one endpoint group: its routes, permissions and operation IDs.

`osapi controller endpoint openapi get` prints the group's part of the OpenAPI
spec.

```bash
osapi controller endpoint get [flags]
```

## Examples

```bash
osapi controller endpoint get --name agent
osapi controller endpoint get --name agent --json
```

## Result

The command prints:

```ansi
$ osapi controller endpoint get --name agent
[1;38;2;0;213;255mEndpoint group:[0m
  [38;2;10;143;176mName:[0m         agent
  [38;2;10;143;176mDescription:[0m  Manage the agents in the fleet: discovery, drain and PKI enrollment

[1;38;2;255;0;255mMethods (2):[0m
  [38;2;176;48;176mNAME   [0m[38;2;176;48;176mVERB  [0m[38;2;176;48;176mPATH                         [0m[38;2;176;48;176mPERMISSION   [0m[38;2;176;48;176mOPERATION ID[0m
  [1mlist[0m   GET   /api/agent                   [38;2;138;138;138magent:read[0m   [38;2;138;138;138mGetAgents[0m
  [1mdrain[0m  POST  /api/agent/[38;2;138;138;138m{hostname}[0m/drain  [38;2;138;138;138magent:write[0m  [38;2;138;138;138mDrainAgent[0m

[1;38;2;255;209;0mTypes (2):[0m
  [38;2;176;143;0mNAME   [0m[38;2;176;143;0mINPUT                     [0m[38;2;176;143;0mOUTPUT[0m
  [1mlist[0m   ListInput                 ListAgentsResponse { items, total_items }
  [1mdrain[0m  DrainInput { hostname* }
```

With `--json`, the command prints:

```json
{
  "name": "agent",
  "doc": "Manage the agents in the fleet: discovery, drain and PKI enrollment",
  "methods": [
    {
      "name": "list",
      "verb": "GET",
      "path": "/api/agent",
      "permission": "agent:read",
      "operation_id": "GetAgents"
    },
    {
      "name": "drain",
      "verb": "POST",
      "path": "/api/agent/{hostname}/drain",
      "permission": "agent:write",
      "operation_id": "DrainAgent"
    }
  ],
  "types": [
    {
      "name": "list",
      "input": "ListInput",
      "output": "ListAgentsResponse { items, total_items }"
    },
    {
      "name": "drain",
      "input": "DrainInput { hostname* }",
      "output": ""
    }
  ]
}
```

## Flags

| Flag     | Description                                  | Default |
| -------- | -------------------------------------------- | ------- |
| `--name` | The endpoint group, such as agent (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
