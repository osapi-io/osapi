# Get

Print how to build a chain, and the building blocks of this binary.

Print what a chain is, how a behavior builds on others and calls them, what
happens when a step fails, and each building block of this binary. Read it
before a behavior requires another, then check the result with
`osapi behavior chain validate`.

```bash
osapi behavior chain spec get
```

## Examples

```bash
osapi behavior chain spec get
osapi behavior chain spec get --json
```

## Result

The command prints:

```ansi
$ osapi behavior chain spec get
[1;38;2;0;213;255mWhat a chain is:[0m
  [1mA chain[0m
    The behaviors that a behavior calls, and the ones that they call: dns
    calls netplan, and netplan calls managedfile. The chain has the name of
    the behavior at its top. It is not declared on its own: Requires on each
    behavior makes it.
    [38;2;138;138;138mUse: osapi behavior chain list shows each chain; osapi behavior chain[0m
    [38;2;138;138;138mget --name <behavior> shows its tree.[0m
```

With `--json`, the command prints:

```json
{
  "kinds": [
    {
      "kind": "chain",
      "title": "What a chain is",
      "terms": [
        {
          "kind": "chain",
          "name": "A chain",
          "does": "The behaviors that a behavior calls, and the ones that they call: dns calls netplan, and netplan calls managedfile. The chain has the name of the behavior at its top. It is not declared on its own: Requires on each behavior makes it.",
          "when": "osapi behavior chain list shows each chain; osapi behavior chain get --name \u003cbehavior\u003e shows its tree."
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

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
