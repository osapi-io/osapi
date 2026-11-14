# Tokens

A token is a JWT signed with `controller.api.security.signing_key`. The
controller refuses a token signed with any other key, so a token made on one
install does not work on another.

```sh
osapi token generate --roles admin --subject ops@example.com --json
osapi token generate --roles read --subject ci --permissions node:read --json
osapi token validate --token <jwt> --json
```

Roles are `admin`, `write` and `read`. `--permissions` grants exactly the
listed permissions instead of a role's; `osapi help token generate --json`
lists them.

The CLI sends `controller.client.security.bearer_token`, usually set as
`OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN`.

A request refused with `unauthorized` has no valid token: missing, expired, or
signed with a different key. One refused with `forbidden` has a valid token
without the permission; `token validate` shows what it carries.
