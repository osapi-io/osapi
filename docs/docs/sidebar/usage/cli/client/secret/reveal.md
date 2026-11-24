# Reveal

Print the value of a key.

The command prints only the value, with no heading and no trace line, so a pipe
gets the value. --json prints &#123;"value": "..."&#125;. The audit log records
each call. A request that uses the secret gives its reference,
secret://&lt;vault&gt;/&lt;key&gt;, not the value.

```bash
osapi client secret reveal [flags]
```

## Examples

```bash
osapi client secret reveal --vault default --key registry-token
osapi client secret reveal --vault default --key registry-token --json
```

## Result

The command prints:

```ansi
$ osapi client secret reveal --vault default --key registry-token
glpat-xxxxxxxxxxxxxxxxxxxx
```

With `--json`, the API's response:

```json
{
  "value": "glpat-xxxxxxxxxxxxxxxxxxxx"
}
```

## Permission

It needs the `secret:reveal` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag      | Description                                            | Default |
| --------- | ------------------------------------------------------ | ------- |
| `--key`   | The name of the key, such as registry-token (required) |         |
| `--vault` | The name of the vault, such as default (required)      |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
