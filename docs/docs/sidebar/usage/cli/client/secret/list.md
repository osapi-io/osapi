# List

List the keys of a vault.

The list shows the metadata of each key, and no value. A request uses a key as
secret://&lt;vault&gt;/&lt;key&gt;.

```bash
osapi client secret list [flags]
```

## Examples

```bash
osapi client secret list --vault default
osapi client secret list --vault default --json
```

## Result

The command prints:

```ansi
$ osapi client secret list --vault default
[1;38;2;0;213;255mKeys (2):[0m
  [38;2;10;143;176mKEY             [0m[38;2;10;143;176mVERSION  [0m[38;2;10;143;176mCHANGED              [0m[38;2;10;143;176mBY[0m
  registry-token  3        2026-01-15 09:12:00  ops@example.com
  db-password     1        2026-01-12 16:40:00  ci@example.com

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "items": [
    {
      "vault": "default",
      "key": "registry-token",
      "version": 3,
      "changed": "2026-01-15T09:12:00Z",
      "by": "ops@example.com"
    },
    {
      "vault": "default",
      "key": "db-password",
      "version": 1,
      "changed": "2026-01-12T16:40:00Z",
      "by": "ci@example.com"
    }
  ],
  "total_items": 2
}
```

## Permission

It needs the `secret:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag      | Description                                       | Default |
| --------- | ------------------------------------------------- | ------- |
| `--vault` | The name of the vault, such as default (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
