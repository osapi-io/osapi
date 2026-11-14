# Validate

Check every behavior's definition against the rules the generated routes,
commands, spec and agent rely on. Exits 1 when anything fails, naming the
behavior, the method and the rule.

```bash
$ osapi behavior validate
[ok] 2 behaviors valid
```

```bash
$ osapi behavior validate --json
{"valid":true,"problems":[]}
```

Each problem in `--json` has `behavior`, `method` (when the problem is in one),
`rule` and `message`. The rules:

| Rule        | Fails when                                                                                                                                                                                        |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`      | the name is not lowercase letters and digits                                                                                                                                                      |
| `placement` | the behavior has no placement                                                                                                                                                                     |
| `doc`       | the behavior or a method has no description                                                                                                                                                       |
| `methods`   | the behavior has no methods                                                                                                                                                                       |
| `duplicate` | two methods share a name                                                                                                                                                                          |
| `verb`      | a method name is a synonym of a standard verb, a read is not `list`, `get` or `status`, or a write is not `create`, `update` or `delete` (action methods such as `restart` are not supported yet) |
| `route`     | two methods would be served at the same route                                                                                                                                                     |
| `key`       | a method's key is not a field of its input                                                                                                                                                        |
| `schema`    | an input or output type has no JSON Schema                                                                                                                                                        |
| `output`    | an output field collides with a field of the per-host result                                                                                                                                      |
| `changed`   | a write's output has no `changed` field                                                                                                                                                           |
| `surface`   | a behavior has CLI without API, is internal but served, or has an input the CLI cannot flag                                                                                                       |
| `example`   | a CLI method has no example, or an example or refusal does not fit its method                                                                                                                     |
| `settings`  | settings are not a struct, or a field lacks the name or description the config needs                                                                                                              |
| `platform`  | no platform has a provider, a platform is declared twice, or a provider misses a method                                                                                                           |
| `requires`  | a required behavior is not registered, or behaviors require one another in a cycle                                                                                                                |
| `secret`    | an input field tagged `secret:"true"` would be recorded on the method's span                                                                                                                      |
