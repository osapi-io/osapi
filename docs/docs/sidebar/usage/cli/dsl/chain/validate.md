# Validate

Check every chain against the chain rules.

It prints each problem with its behavior, method, rule and message, then the fix
of each rule, and exits 1 when a rule fails. `osapi dsl behavior validate`
checks the rules that hold when the behaviors are built, and `go test` checks
the ones that need a run of the examples.

The rules:

- `requires`: Each required behavior is registered, and no behavior requires
  itself through others. Fix: Import the required behavior in
  internal/behaviors/all, or remove the requirement that makes the cycle.
- `internal`: Another behavior requires each internal behavior. Nothing else can
  call it, because it has no API and no CLI. Fix: Add behavior.Requires(it) to
  the behavior that uses it, or remove behavior.Internal if an operator calls
  it.
- `called`: A provider calls only the behaviors that its behavior requires. The
  agent builds the provider of each required behavior, and no other. Fix: Add
  the behavior that it calls to behavior.Requires, and read its provider with
  host.Require.
- `used`: An example calls each behavior that a behavior requires. A requirement
  that no example calls is not tested, or is not needed. Fix: Add an example
  whose provider calls the required behavior, or remove the requirement.
- `secret`: A call gives a secret of its input only to a secret.Value field of
  the behavior that it calls. A field of another type, such as a string, records
  the value in its chain record, its span or an audit entry. The examples run
  each chain, so their tests check it. Fix: Give the field of the called
  behavior the type secret.Value, and read it there with Reveal.

```bash
osapi dsl chain validate
```

## Examples

```bash
osapi dsl chain validate
osapi dsl chain validate --json
```

## Result

The command prints:

```ansi
$ osapi dsl chain validate
[1;38;2;0;213;255mValidation:[0m
  [38;2;10;143;176mBehaviors:[0m  33
  [38;2;10;143;176mValid:[0m      yes
```

With `--json`, the command prints:

```json
{
  "behaviors": 33,
  "valid": true
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
