# Get

Print the Go SDK: its packages, how a call works, its types and its services.

The list comes from this binary, from types.All and client.Services. Thus it
agrees with the build that you run. The SDK skill is made from it.

```bash
osapi sdk spec get
```

## Examples

```bash
osapi sdk spec get
osapi sdk spec get --json
```

## Result

The command prints:

```ansi
$ osapi sdk spec get
[1;38;2;0;213;255mPackages:[0m
  [1mgithub.com/osapi-io/osapi/pkg/sdk/client[0m
    Calls the API of the controller: client.New(url, token) is a Client, and
    each service of the API is a field of it, such as c.Hostname.
    [38;2;138;138;138mUse: In a program that calls osapi. The services below are its fields.[0m

  [1mgithub.com/osapi-io/osapi/pkg/sdk/types[0m
    The vocabulary of the API: the permissions, the roles, the targets and
    each closed set of values, as types with constants.
    [38;2;138;138;138mUse: For each value of a closed set: a constant such as[0m
    [38;2;138;138;138mtypes.ContentRaw, never a string.[0m
```

With `--json`, the command prints:

```json
{
  "kinds": [
    {
      "kind": "package",
      "title": "Packages",
      "terms": [
        {
          "kind": "package",
          "name": "github.com/osapi-io/osapi/pkg/sdk/client",
          "does": "Calls the API of the controller: client.New(url, token) is a Client, and each service of the API is a field of it, such as c.Hostname.",
          "when": "In a program that calls osapi. The services below are its fields."
        },
        {
          "kind": "package",
          "name": "github.com/osapi-io/osapi/pkg/sdk/types",
          "does": "The vocabulary of the API: the permissions, the roles, the targets and each closed set of values, as types with constants.",
          "when": "For each value of a closed set: a constant such as types.ContentRaw, never a string."
        }
      ]
    }
  ]
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
