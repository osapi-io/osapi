# Tokens

A token is a JWT signed with the first key of
`controller.api.security.signing_keys`, with that key's `id` as its `kid`. The
controller verifies a token with the key that its `kid` names, and refuses a
token with no `kid` or an unknown one, so a token made on one install does not
work on another, and a token stops when its key leaves the list.
`token validate` shows the `kid`.

```sh
osapi controller token generate --roles admin --subject ops@example.com --json
osapi controller token generate --roles read --subject ci --permissions status:read --json
osapi controller token validate --token <jwt> --json
```

Roles are `admin`, `operator`, `automation`, `auditor`, `read` and `observer`;
`osapi controller role get --name <role>` lists each one's permissions.
`--permissions` grants exactly the listed permissions instead of a role's;
`osapi help token generate --json` lists them.

The CLI sends `controller.client.security.bearer_token`, usually set as
`OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN`.

A request refused with `unauthorized` has no valid token: missing, expired, or
signed with a different key. One refused with `forbidden` has a valid token
without the permission; `controller token validate` shows what it carries.
