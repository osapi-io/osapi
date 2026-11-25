# Get

Get the catalog of the controller.

The catalog is what the UI reads after sign-in: each behavior with its methods,
the areas, the permissions, the built-in roles, the permission of each
operation, the tones of the statuses and the counts, and the host families. The
controller builds it from its registries.

```bash
osapi client catalog get
```

## Examples

```bash
osapi client catalog get
osapi client catalog get --json
```

## Result

The command prints:

```ansi
$ osapi client catalog get
[1;38;2;0;213;255mCatalog:[0m
  [38;2;10;143;176mHost families:[0m  debian, darwin
  [38;2;10;143;176mRender tags:[0m
  [38;2;10;143;176mOutcomes:[0m

[1;38;2;255;0;255mBehaviors (1):[0m
  [38;2;176;48;176mNAME      [0m[38;2;176;48;176mTITLE     [0m[38;2;176;48;176mAREA  [0m[38;2;176;48;176mINTERNAL  [0m[38;2;176;48;176mPLATFORMS       [0m[38;2;176;48;176mMETHODS[0m
  [1mhostname[0m  Hostname  node  no        debian, darwin  get

[1;38;2;255;209;0mAreas (1):[0m
  [38;2;176;143;0mNAME  [0m[38;2;176;143;0mLABEL[0m
  [1mnode[0m  Node

[1;38;2;69;174;23mRoles (1):[0m
  [38;2;58;143;20mNAME      [0m[38;2;58;143;20mDESCRIPTION[0m
  [1mobserver[0m  A dashboard or a monitor.

[1;38;2;107;92;255mOperations (1):[0m
  [38;2;89;76;216mID               [0m[38;2;89;76;216mPERMISSION[0m
  [38;2;138;138;138mGetNodeHostname[0m  [38;2;138;138;138mhostname:read[0m

[1;38;2;0;213;255mTargets (0):[0m
  [38;2;138;138;138mNone.[0m

[1;38;2;255;0;255mController (0):[0m
  [38;2;138;138;138mNone.[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "behaviors": [
    {
      "name": "hostname",
      "title": "Hostname",
      "area": "node",
      "doc": "The hostname of a host.",
      "internal": false,
      "platforms": ["debian", "darwin"],
      "uses": [],
      "used_by": [],
      "chain": "",
      "methods": [
        {
          "name": "get",
          "kind": "read",
          "doc": "Get the hostname.",
          "http_method": "GET",
          "path": "/api/node/{hostname}/hostname",
          "permission": "hostname:read",
          "platforms": ["debian", "darwin"],
          "runs_on": ["debian", "darwin"],
          "input": "GetInput {}",
          "output": "GetOutput { hostname }",
          "fields": [],
          "columns": [
            {
              "name": "hostname",
              "label": "Hostname",
              "doc": "The hostname.",
              "kind": "value"
            }
          ],
          "calls": [],
          "examples": []
        }
      ]
    }
  ],
  "areas": [
    {
      "name": "node",
      "label": "Node"
    }
  ],
  "permissions": ["hostname:read", "hostname:write"],
  "roles": [
    {
      "name": "observer",
      "description": "A dashboard or a monitor.",
      "permissions": ["health:read"]
    }
  ],
  "operations": [
    {
      "id": "GetNodeHostname",
      "permission": "hostname:read"
    }
  ],
  "status_tones": {
    "failed": "err",
    "ok": "ok"
  },
  "count_tones": {
    "failed": "err"
  },
  "host_families": ["debian", "darwin"],
  "targets": null,
  "controller": null,
  "render_tags": null,
  "outcomes": null
}
```

## Permission

It needs the `catalog:read` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Notes

The table leaves out the permissions, the tones and the doc of each behavior,
and the permissions of each role. --json prints the whole catalog, with the
methods of each behavior. osapi controller role get --name &lt;role&gt; shows
the permissions of a role.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
