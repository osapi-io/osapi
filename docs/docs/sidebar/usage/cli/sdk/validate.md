# Validate

Check the SDK against its rules.

It prints each problem with its name, rule and message, then the fix of each
rule, and exits 1 when a rule fails.

The rules:

- `doc`: Each type, each service and each method of the SDK says what it is.
  Fix: Give the type a doc comment, or the schema or the operation of the spec a
  description, and run just generate.
- `values`: Each type for values has at least one value. Fix: Give the type
  constants and list them in its Values.
- `example`: Each method of the client has an example, which examples/sdk/client
  runs, but a method that a program calls, which the spec marks x-osapi-program.
  Fix: Add an Example to the method's definition, and run just generate.
- `client`: Each service that client.Services lists is a field of client.Client,
  and each of its methods is a method of that field. Fix: Run just generate:
  client.Services and the client come from one run.

```bash
osapi sdk validate
```

## Examples

```bash
osapi sdk validate
osapi sdk validate --json
```

## Result

The command prints:

```ansi
$ osapi sdk validate
[1;38;2;0;213;255mValidation:[0m
  [38;2;10;143;176mTypes:[0m     22
  [38;2;10;143;176mServices:[0m  1
  [38;2;10;143;176mValid:[0m     yes
```

With `--json`, the command prints:

```json
{
  "types": 22,
  "services": 1,
  "valid": true
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
