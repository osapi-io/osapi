# List

List each behavior in a chain: what it calls, and what calls it.

\*\*Chain\*\* is the tree of the behaviors that a behavior calls, down to the
last one. A building block, with `Internal`, is in the list for the behaviors
that call it.

```bash
osapi dsl chain list
```

## Examples

```bash
osapi dsl chain list
osapi dsl chain list --json
```

## Result

The command prints:

```ansi
$ osapi dsl chain list
[1;38;2;0;213;255mBehaviors in a chain (3):[0m
  [38;2;10;143;176mBEHAVIOR     [0m[38;2;10;143;176mINTERNAL  [0m[38;2;10;143;176mCHAIN                        [0m[38;2;10;143;176mCALLS        [0m[38;2;10;143;176mUSED BY[0m
  dns                    dns → netplan → managedfile  netplan      none
  netplan      yes       netplan → managedfile        managedfile  route, dns, interface
  managedfile  yes       managedfile                  none         certificate, file, netplan
```

With `--json`, the command prints:

```json
{
  "items": [
    {
      "behavior": "dns",
      "chain": "dns → netplan → managedfile",
      "calls": ["netplan"],
      "used_by": []
    },
    {
      "behavior": "netplan",
      "internal": true,
      "chain": "netplan → managedfile",
      "calls": ["managedfile"],
      "used_by": ["route", "dns", "interface"]
    },
    {
      "behavior": "managedfile",
      "internal": true,
      "chain": "managedfile",
      "calls": [],
      "used_by": ["certificate", "file", "netplan"]
    }
  ]
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
