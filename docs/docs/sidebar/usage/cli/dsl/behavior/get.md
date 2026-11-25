# Get

Show a behavior: its surfaces, chain, methods, types and telemetry.

It shows where osapi serves the behavior, its platforms, the programs that it
runs with sudo, the behaviors that it uses and that use it, its methods with
their routes and permissions, its types, its settings and its telemetry.

\*\*Uses\*\* lists the behaviors this one is built on. The agent builds their
providers first, and the methods of this behavior call theirs on the same host,
in the same job. If a behavior that it uses cannot run on a platform, this one
reports `skipped` there too. \*\*Used by\*\* lists the behaviors built on this
one. A behavior with `Internal: yes` has no route, no command and no permission.
Only the behaviors under \*\*Used by\*\* call it.
`osapi dsl behavior method get` shows one method in full: its fields with their
constraints, its output and its examples.

```bash
osapi dsl behavior get [flags]
```

## Examples

```bash
osapi dsl behavior get --name sysctl
osapi dsl behavior get --name sysctl --json
```

## Result

The command prints:

```ansi
$ osapi dsl behavior get --name sysctl
[1;38;2;0;213;255mBehavior:[0m
  [38;2;10;143;176mName:[0m         sysctl
  [38;2;10;143;176mDescription:[0m  Kernel parameters, written to sysctl.d files and applied
  [38;2;10;143;176mPlacement:[0m    node
  [38;2;10;143;176mInternal:[0m     no
  [38;2;10;143;176mAPI:[0m          yes
  [38;2;10;143;176mCLI:[0m          yes
  [38;2;10;143;176mNext:[0m         no
  [38;2;10;143;176mPlatforms:[0m    debian
  [38;2;10;143;176mSudo:[0m         sysctl
  [38;2;10;143;176mRuns:[0m         none
  [38;2;10;143;176mWrites:[0m       /etc/sysctl.d/osapi-*.conf
  [38;2;10;143;176mKeeps state:[0m  no
  [38;2;10;143;176mUses:[0m         managedfile
  [38;2;10;143;176mUsed by:[0m      none

[1;38;2;255;0;255mMethods (2):[0m
  [38;2;176;48;176mNAME    [0m[38;2;176;48;176mVERB   [0m[38;2;176;48;176mHTTP  [0m[38;2;176;48;176mROUTE                              [0m[38;2;176;48;176mPERMISSION    [0m[38;2;176;48;176mDESCRIPTION[0m
  [1mget[0m     read   GET   /api/node/[38;2;138;138;138m{hostname}[0m/sysctl/[38;2;138;138;138m{key}[0m  [38;2;138;138;138msysctl:read[0m   Get a kernel parameter's current value, and whether osapi manages it
  [1mupdate[0m  [1mwrite[0m  PUT   /api/node/[38;2;138;138;138m{hostname}[0m/sysctl/[38;2;138;138;138m{key}[0m  [38;2;138;138;138msysctl:write[0m  Set a kernel parameter, manage it, and apply it

[1;38;2;255;209;0mTypes (2):[0m
  [38;2;176;143;0mNAME    [0m[38;2;176;143;0mINPUT                         [0m[38;2;176;143;0mOUTPUT[0m
  [1mget[0m     GetInput { key* }             Entry { key, value, managed }
  [1mupdate[0m  UpdateInput { key*, value* }  MutationResult { key, changed }

[1;38;2;69;174;23mSettings (1):[0m
  [38;2;58;143;20mKEY                   [0m[38;2;58;143;20mAGENT FLAG    [0m[38;2;58;143;20mTYPE    [0m[38;2;58;143;20mDEFAULT[0m
  behaviors.sysctl.dir  --sysctl-dir  [1mstring[0m  /etc/sysctl.d

[1;38;2;107;92;255mTelemetry:[0m
  [38;2;89;76;216mSpan:[0m             behavior.sysctl.<method>
  [38;2;89;76;216mChild spans:[0m      behavior.managedfile.<method>, exec sysctl
  [38;2;89;76;216mMetrics:[0m          osapi_behavior_calls_total, osapi_behavior_duration_seconds
  [38;2;89;76;216mAttributes:[0m       osapi.behavior.name, osapi.behavior.method
  [38;2;89;76;216mRecorded inputs:[0m  key (get, update)
```

With `--json`, the command prints:

```json
{
  "name": "sysctl",
  "doc": "Kernel parameters, written to sysctl.d files and applied",
  "placement": "node",
  "internal": false,
  "api": true,
  "cli": true,
  "next": false,
  "platforms": ["debian"],
  "sudo": ["sysctl"],
  "runs": null,
  "writes": ["/etc/sysctl.d/osapi-*.conf"],
  "keeps_state": false,
  "uses": ["managedfile"],
  "used_by": null,
  "methods": [
    {
      "name": "get",
      "verb": "read",
      "http_method": "GET",
      "route": "/api/node/{hostname}/sysctl/{key}",
      "permission": "sysctl:read",
      "command": "osapi client node sysctl get",
      "doc": "Get a kernel parameter's current value, and whether osapi manages it"
    },
    {
      "name": "update",
      "verb": "write",
      "http_method": "PUT",
      "route": "/api/node/{hostname}/sysctl/{key}",
      "permission": "sysctl:write",
      "command": "osapi client node sysctl update",
      "doc": "Set a kernel parameter, manage it, and apply it"
    }
  ],
  "types": [
    {
      "name": "get",
      "input": "GetInput { key* }",
      "output": "Entry { key, value, managed }"
    },
    {
      "name": "update",
      "input": "UpdateInput { key*, value* }",
      "output": "MutationResult { key, changed }"
    }
  ],
  "settings": [
    {
      "key": "behaviors.sysctl.dir",
      "flag": "sysctl-dir",
      "type": "string",
      "default": "/etc/sysctl.d"
    }
  ],
  "telemetry": {
    "span": "behavior.sysctl.\u003cmethod\u003e",
    "child_spans": ["behavior.managedfile.\u003cmethod\u003e", "exec sysctl"],
    "metrics": [
      "osapi_behavior_calls_total",
      "osapi_behavior_duration_seconds"
    ],
    "attributes": ["osapi.behavior.name", "osapi.behavior.method"],
    "recorded": "key (get, update)",
    "inputs": {
      "get": ["key"],
      "update": ["key"]
    }
  }
}
```

## Flags

| Flag     | Description                                           | Default |
| -------- | ----------------------------------------------------- | ------- |
| `--name` | The behavior, from osapi dsl behavior list (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
