# Validate

Check every endpoint group against the rules of a definition.

It prints each problem with its group, method, rule and message, then the fix of
each rule, and exits 1 when a rule fails. `osapi dsl endpoint spec get`
describes each part of a group.

The rules:

- `example`: Each flag that an Example names is a flag of the method's command:
  a field of its input, a field of the struct that in:"body" names, or --json,
  --debug or --osapi-file. Fix: Spell each flag of the example as the command
  has it: osapi dsl endpoint get --name &lt;group&gt; lists the methods, and the
  command's --help lists its flags.
- `prints`: The output that Prints gives an example is a value of the method's
  output type, and Prints follows an Example. Fix: Give Prints a value of the
  method's output type, after the Example that it shows.
- `in`: Each in tag of an input field is path, query, form or body. Fix: Correct
  the value. A field with another value is read from the request body, not from
  where the tag says.
- `reason`: Each error that a method maps has a reason of endpoint.Reasons. Fix:
  Use a Reason constant, such as endpoint.ReasonNotManaged, in Because.
- `field`: The field that Location names is a json field of the method's output,
  and the field that Example\[T] names is a json field of T. Fix: Correct the
  name to the json name of the field, such as job_id.
- `area`: The group has an area, one of area.All, so the UI, the SDK's pages and
  the docs site can place its operations. Fix: Add endpoint.Area(...) to the
  group's definition with a constant of package area, such as area.Fleet. Add a
  new area to area.All with its label first.
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
  endpoint.Uses\[Deps]() to Define.
- `subdoc`: Each Sub segment that makes a command of its own, such as trace in
  osapi client telemetry trace get, has a SubDoc: the one line of help of that
  command. Fix: Add endpoint.SubDoc("segment", "...") to Define, with what the
  segment's commands do, such as Read a trace from the tracing backend.
- `validate`: Each input field has a validate tag: its rules, or omitempty for
  an optional field with none. The spec carries the whole tag, and the server
  answers 400 to a value that breaks it. Fix: Add validate:"..." to the field,
  such as validate:"required,min=1,max=255", or validate:"omitempty" for an
  optional field with no other rule.
- `fails`: Each rule of an input field has a .Fails input that breaks it, which
  the API refuses with 400. An endpoint's required needs one, with the field
  empty; a behavior's needs none, as its generated tests leave the field out. A
  path field's required, and a min of 1 on a string that is required, optional
  or in the path, need none: their only bad value is empty. Fix: Add
  .Fails("&lt;what is wrong&gt;", &lt;Input&gt;&#123;...&#125;) to the method,
  with the other fields valid and the field's value breaking the rule that the
  message names, such as a name of 256 characters for max=255.
- `optional`: An optional field of a write's request body, with omitempty in its
  validate tag, has omitempty in its json tag too, so a client leaves it out
  when it is empty instead of sending "" or 0, which its other rules can refuse.
  Fix: Add ,omitempty to the field's json tag, such as json:"reason,omitempty".
- `schema`: The spec builds: each input and output type is one that JSON can
  encode, and each validate rule has a schema. Fix: Use structs, slices, maps,
  strings, numbers, bools and time.Time; not channels or functions. Register the
  schema of a custom validate rule with rule.Register, beside its validator in
  internal/validation.

```bash
osapi dsl endpoint validate
```

## Examples

```bash
osapi dsl endpoint validate
osapi dsl endpoint validate --json
```

## Result

The command prints:

```ansi
$ osapi dsl endpoint validate
[1;38;2;0;213;255mValidation:[0m
  [38;2;10;143;176mGroups:[0m  10
  [38;2;10;143;176mValid:[0m   yes
```

With `--json`, the command prints:

```json
{
  "groups": 10,
  "valid": true
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
