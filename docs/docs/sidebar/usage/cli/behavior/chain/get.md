# Get

Show a chain: its tree, and what each call of each example sends and gets.

It runs each example of the behavior on the fake host that `go test` uses, and
records each `behavior.Call`: the behavior, the method, the input, and the
output or the error. The table cuts each input and output; `--json` has them in
full, under `runs`. An input field tagged `secret` shows as `[redacted]`.

```bash
osapi behavior chain get [flags]
```

## Examples

```bash
osapi behavior chain get --name dns
osapi behavior chain get --name dns --json
```

## Result

The command prints:

```ansi
$ osapi behavior chain get --name dns
[1;38;2;0;213;255mChain:[0m
  [38;2;176;48;176mBehavior:[0m  dns
  [38;2;176;48;176mChain:[0m     dns → netplan → managedfile
  [38;2;176;48;176mCalls:[0m     netplan
  [38;2;176;48;176mUsed by:[0m   none

[1;38;2;255;0;255mSet the servers of an interface (update):[0m
  [38;2;176;48;176mCALL                     [0m[38;2;176;48;176mSENDS                                                      [0m[38;2;176;48;176mGETS[0m
  dns update               interface_name=osapi0 servers=[9.9.9.9]                    changed=true interface_name=osapi0
  [38;2;138;138;138m└─ [0mnetplan get           name=osapi0                                                id=osapi0 name=osapi0 section=ethernets type=ethernet
  [38;2;138;138;138m└─ [0mnetplan update        content=<9 lines> path=/etc/netplan/osapi-dns-osapi0.yaml  changed=true path=/etc/netplan/osapi-dns-osapi0.yaml
  · └─ managedfile update  content=<9 lines> path=/etc/netplan/osapi-dns-osapi0.yaml  changed=true path=/etc/netplan/osapi-dns-osapi0.yaml
```

With `--json`, the command prints:

```json
{
  "behavior": "dns",
  "chain": "dns → netplan → managedfile",
  "calls": ["netplan"],
  "used_by": [],
  "examples": [
    {
      "example": "Set the servers of an interface (update)",
      "calls": [
        {
          "call": "dns update",
          "sends": "interface_name=osapi0 servers=[9.9.9.9]",
          "gets": "changed=true interface_name=osapi0"
        },
        {
          "call": "└─ netplan get",
          "sends": "name=osapi0",
          "gets": "id=osapi0 name=osapi0 section=ethernets type=ethernet"
        },
        {
          "call": "└─ netplan update",
          "sends": "content=\u003c9 lines\u003e path=/etc/netplan/osapi-dns-osapi0.yaml",
          "gets": "changed=true path=/etc/netplan/osapi-dns-osapi0.yaml"
        },
        {
          "call": "· └─ managedfile update",
          "sends": "content=\u003c9 lines\u003e path=/etc/netplan/osapi-dns-osapi0.yaml",
          "gets": "changed=true path=/etc/netplan/osapi-dns-osapi0.yaml"
        }
      ]
    }
  ],
  "runs": null
}
```

## Flags

| Flag     | Description                                             | Default |
| -------- | ------------------------------------------------------- | ------- |
| `--name` | The behavior, from osapi behavior chain list (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
