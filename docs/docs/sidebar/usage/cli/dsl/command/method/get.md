# Get

Show one method of a local group: its command, flags, output, examples and
refusals.

Each flag has the validate tag of its field. `--json` has the JSON Schema of the
output. `osapi dsl behavior method get` and `osapi dsl endpoint method get` show
a method of a behavior and of an endpoint group the same way.

```bash
osapi dsl command method get [flags]
```

## Examples

```bash
osapi dsl command method get --name "controller token" --method generate
osapi dsl command method get --name "controller token" --method generate --json
```

## Result

The command prints:

```ansi
$ osapi dsl command method get --name "controller token" --method generate
[1;38;2;0;213;255mMethod:[0m
  [38;2;10;143;176mCommand:[0m      osapi controller token generate
  [38;2;10;143;176mMethod:[0m       generate
  [38;2;10;143;176mDescription:[0m  Generate a new token
  [38;2;10;143;176mProcess:[0m      no
  [38;2;10;143;176mOutput type:[0m  GenerateOutput
  [38;2;10;143;176mRefuses:[0m      no roles, no subject

[1;38;2;255;0;255mFlags (2):[0m
  [38;2;176;48;176mFLAG       [0m[38;2;176;48;176mSHORT  [0m[38;2;176;48;176mTYPE      [0m[38;2;176;48;176mVALIDATE  [0m[38;2;176;48;176mHELP[0m
  --roles    r      [1m[]string[0m  required  Roles for the token: built-in, or a custom role of the config
  --subject  u      [1mstring[0m    required  Subject for the token (e.g., user ID or unique identifier)

[1;38;2;255;209;0mOutput (3):[0m
  [38;2;176;143;0mFIELD    [0m[38;2;176;143;0mTYPE[0m
  token    [1mstring[0m
  subject  [1mstring[0m
  roles    [1marray[0m

[1;38;2;69;174;23mExamples (1):[0m
  [38;2;58;143;20mFLAGS[0m
  --subject user123 --roles=read
```

With `--json`, the command prints:

```json
{
  "command": "osapi controller token generate",
  "method": "generate",
  "doc": "Generate a new token",
  "process": false,
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
  ],
  "output_type": "GenerateOutput",
  "results": [
    {
      "field": "token",
      "type": "string",
      "description": ""
    },
    {
      "field": "subject",
      "type": "string",
      "description": ""
    },
    {
      "field": "roles",
      "type": "array",
      "description": ""
    }
  ],
  "examples": [
    {
      "flags": "--subject user123 --roles=read",
      "doc": ""
    }
  ],
  "refuses": ["no roles", "no subject"]
}
```

## Flags

| Flag       | Description                                                                  | Default |
| ---------- | ---------------------------------------------------------------------------- | ------- |
| `--method` | The method, such as list or roles get, from osapi dsl command get (required) |         |
| `--name`   | The command path of the local group, such as controller token (required)     |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
