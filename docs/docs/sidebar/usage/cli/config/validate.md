# Validate

Check a config file before osapi starts with it.

The command starts nothing. It exits 1 when the config is not valid, and lists
each problem with its rule, the key and the message, then the fix of each rule.
Each problem in `--json` has `rule`, `key` (when the problem is about one key),
`message` and `fix`. `osapi config list` shows each key, with its type, rule and
default.

The rules:

- `file`: The config file exists, and osapi can read it. Fix: Give the path with
  -f, or put the file at /etc/osapi/osapi.yaml. Check that the user that runs
  osapi can read it.
- `parse`: The file is YAML, and each value has the type of its key. Fix:
  Correct the YAML at the line that the message names. osapi config list shows
  the type of each key.
- `value`: Each value passes the validate rule of its key, with the defaults and
  the environment applied. Fix: Set the key to a value that its rule allows.
  osapi config get --key &lt;key&gt; shows its rule, type and default.
- `setting`: Each setting of a behavior, under behaviors.&lt;name&gt;, passes
  its validate rule. Fix: Set the key to a value that its rule allows. osapi
  behavior get --name &lt;name&gt; lists the settings of the behavior.
- `bucket`: Each NATS bucket has its own name. A shared bucket mixes the keys of
  two roles. Fix: Give one of the two keys that the message names a different
  bucket.

```bash
osapi config validate
```

## Examples

```bash
osapi config validate
osapi config validate -f ./osapi.yaml --json
```

## Result

The command prints:

```ansi
$ osapi config validate
[1;38;2;0;213;255mConfig:[0m
  [38;2;10;143;176mFile:[0m   /etc/osapi/osapi.yaml
  [38;2;10;143;176mValid:[0m  yes

[1;38;2;255;0;255mProblems (0):[0m
  [38;2;138;138;138mNone.[0m
```

With `--json`, the command prints:

```json
{
  "file": "/etc/osapi/osapi.yaml",
  "valid": true,
  "problems": []
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
