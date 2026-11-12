# Errors

Under `--json`, a failure writes one document to stderr:

```json
{"error": "insufficient permissions", "code": "forbidden", "status": 403}
```

| `code`                 | What to do                                                                           |
| ---------------------- | ------------------------------------------------------------------------------------ |
| `usage_error` (exit 2) | the command, a flag or a required flag is wrong; run `osapi help <command> --json`   |
| `invalid_request`      | an input breaks its schema; run `osapi behavior describe <behavior> <method> --json` |
| `unauthorized`         | the token is missing or invalid                                                      |
| `forbidden`            | the token lacks the method's `permission`                                            |
| `not_found`            | the resource or host does not exist                                                  |
| `server_error`         | the controller failed; retrying may help                                             |
| `fatal`                | the CLI could not start, usually its config file; `details` says which               |

A host that failed inside a successful command is not here: it is a result row
with `status: failed` and its own `error`.
