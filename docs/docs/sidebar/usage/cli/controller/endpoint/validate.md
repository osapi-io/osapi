# Validate

Check every endpoint group against the rules of a definition.

It exits 1 when a rule fails, and names the group, the method, the rule, what is
wrong and the fix. Each problem in `--json` has `group`, `method` (when the
problem is in one), `rule`, `message` and `fix`.
`osapi controller endpoint spec` describes each part of a group.

The rules:

- `name`: The group and each method have a name of lowercase letters and digits,
  starting with a letter. Fix: Name the group as its path segment, such as
  agent, and each method for what it does, such as list or drain.
- `methods`: The group has at least one method. Fix: Add a method with
  endpoint.Get, Post, Put or Delete.
- `unique`: No two methods of a group have one name. Fix: Give each method its
  own name.
- `segment`: Each Sub and Action segment is lowercase letters and digits. Fix:
  Use one word, such as pending or drain.
- `action`: A method with an Action has a key: an input field tagged in:"path".
  Fix: Tag the field the action is on in:"path", or use Sub for a fixed path.
- `key`: An input has at most one field tagged in:"path". Fix: Tag one field
  in:"path"; a group has one key.
- `id`: Each operationId is used once, and starts with a capital letter, as each
  operationId of the API does. Fix: Give the method its own ID with .ID, in
  Pascal case, such as GetAgents.
- `route`: No two methods have one verb and path. Fix: Change the verb, the Sub
  or the Action.
- `handler`: A group that Uses a type gives each method a handler,
  func(context.Context, D, In) (Out, error), with D that type and In and Out the
  method's types. A group with a handler Uses a type. Fix: Add .Handle(fn) to
  each method, with the signature that the message names, and
  endpoint.Uses[Deps](<>) to Define.
- `schema`: The spec builds: each input and output type is one that JSON can
  encode. Fix: Use structs, slices, maps, strings, numbers, bools and time.Time;
  not channels or functions.

```bash
osapi controller endpoint validate
```

## Examples

```bash
osapi controller endpoint validate
osapi controller endpoint validate --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |
| `-p, --port`       | Port the server will bind to | `8080`                  |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
