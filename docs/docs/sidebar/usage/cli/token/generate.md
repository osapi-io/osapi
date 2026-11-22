# Generate

Generate a new token.

```bash
osapi token generate [flags]
```

## Examples

```bash
osapi token generate --subject user123 --roles=read
osapi token generate --subject svc@example.com --roles=read \
  -p node:read -p health:read
```

## Flags

| Flag                | Description                                                                                     | Default |
| ------------------- | ----------------------------------------------------------------------------------------------- | ------- |
| `-p, --permissions` | Direct permissions, which override the roles; osapi token roles get --name admin lists each one |         |
| `-r, --roles`       | Roles for the token: built-in, or a custom role of the config (required)                        |         |
| `-u, --subject`     | Subject for the token (e.g., user ID or unique identifier) (required)                           |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
