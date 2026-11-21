# Validate

Check every chain against the chain rules.

It prints each problem with its behavior, method, rule and message, then the fix
of each rule, and exits 1 when a rule fails. `osapi behavior validate` checks
the rules that hold when the behaviors are built, and `go test` checks the ones
that need a run of the examples.

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

```bash
osapi behavior chain validate
```

## Examples

```bash
osapi behavior chain validate
osapi behavior chain validate --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
