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

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
