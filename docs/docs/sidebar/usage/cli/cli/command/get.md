# Get

Describe one group: its commands, and the flags and refusals of a local group.

`--name` is the command path of the group, such as `token` or `agent`: a local
group, or the group of an endpoint or a behavior. A command of a local group
also has its flags, each with the validate tag of its field, and the inputs that
its `.Fails` declares, which `TestRefusals` runs.

```bash
osapi cli command get [flags]
```

## Examples

```bash
osapi cli command get --name token
osapi cli command get --name agent --json
```

## Result

The command prints:

```ansi
$ osapi cli command get --name token
[1;38;2;0;213;255mGroup:[0m
  [38;2;176;48;176mName:[0m         token
  [38;2;176;48;176mDescription:[0m  Generate and validate API tokens

[1;38;2;255;0;255mosapi token generate:[0m
  [38;2;176;48;176mMethod:[0m       generate
  [38;2;176;48;176mDescription:[0m  Generate a new token
  [38;2;176;48;176mRefuses:[0m      no roles, no subject

[1;38;2;255;209;0mFlags (2):[0m
  [38;2;176;143;0mFLAG       [0m[38;2;176;143;0mSHORT  [0m[38;2;176;143;0mTYPE      [0m[38;2;176;143;0mVALIDATE  [0m[38;2;176;143;0mHELP[0m
  --roles    r      [1m[]string[0m  required  Roles for the token: built-in, or a custom role of the config
  --subject  u      [1mstring[0m    required  Subject for the token (e.g., user ID or unique identifier)

[1;38;2;69;174;23mosapi token roles list:[0m
  [38;2;58;143;20mMethod:[0m       roles list
  [38;2;58;143;20mDescription:[0m  List the roles and the permissions each grants
```

With `--json`, the command prints:

```json
{
  "name": "token",
  "doc": "Generate and validate API tokens",
  "commands": [
    {
      "command": "osapi token generate",
      "method": "generate",
      "short": "Generate a new token",
      "fails": ["no roles", "no subject"],
      "flags": [
        {
          "flag": "--roles",
          "short": "r",
          "type": "[]string",
          "validate": "required",
          "doc": "Roles for the token: built-in, or a custom role of the config"
        },
        {
          "flag": "--subject",
          "short": "u",
          "type": "string",
          "validate": "required",
          "doc": "Subject for the token (e.g., user ID or unique identifier)"
        }
      ]
    },
    {
      "command": "osapi token roles list",
      "method": "roles list",
      "short": "List the roles and the permissions each grants"
    }
  ]
}
```

## Flags

| Flag     | Description                                                      | Default |
| -------- | ---------------------------------------------------------------- | ------- |
| `--name` | The command path of the group, such as token or agent (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
