# Validate

Check every local group against the rules of a definition, and the output of
every group.

It prints each problem with its group, method, rule and message, then the fix of
each rule, and exits 1 when a rule fails. `osapi cli command spec get` describes
each part of a group.

The rules:

- `name`: The group, each word of Under, each method and each Sub have a name of
  lowercase letters and digits, starting with a letter. Fix: Name the group for
  what its commands are about, such as token, and each method for what it does,
  such as generate.
- `methods`: The group has at least one method. Fix: Add a method with
  command.Method, with its input and output types and its name.
- `unique`: No two methods of a group have one name. Fix: Give each method its
  own name.
- `verb`: The last word of each command is a verb (list, get, create, update,
  delete) or an action in command.Actions, as TestCommandShape holds each
  command to. A command directly under the root, of a Flat group, has a word of
  command.Roots instead. Fix: Name the method for a verb, and put a noun before
  it with .Sub, such as roles list. Add a word to command.Actions only for a
  real action, never for a noun, and to command.Roots only for a command that
  belongs at the root.
- `help`: Each method has a Short, and a method with input has an Example. Each
  flag that an Example names is a flag of the command: a field of its input, a
  key that its Flag names, or --json, --debug or --osapi-file. Fix: Add
  .Short("&lt;what it does&gt;") and .Example("--flag value") to the method, and
  spell each flag of an example as the command has it.
- `flag`: Each value is a flag: the input is a struct, each field has a json
  name that is its flag with _ as -, no flag or in tag, and a type a flag holds
  (string, bool, int, \[]string, map\[string]string or \[]byte). A command takes
  no positional argument. Fix: Make the input a struct, tag each field
  json:"name", drop the flag and in tags, and use one of the flag types.
- `doc`: Each input field has a doc tag: the help of its flag. Fix: Add
  doc:"..." to the field: a capital letter first and no period at the end.
- `validate`: Each input field has a validate tag: its rules, or omitempty for
  an optional field with none. Run checks the value against it before the
  handler runs, and a value that breaks it is a usage error. Fix: Add
  validate:"..." to the field, such as validate:"required,min=1,max=255", or
  validate:"omitempty" for an optional field with no other rule.
- `fails`: Each rule of an input field has a .Fails input that breaks it, which
  the command refuses with a usage error, required included. A min of 1 on a
  string that is required or optional needs none: its only bad value is empty.
  Fix: Add .Fails("&lt;what is wrong&gt;", &lt;Input&gt;&#123;...&#125;) to the
  method, with the other fields valid and the field's value breaking the rule
  that the message names, such as a name of 256 characters for max=255.
- `subdoc`: Each Sub segment that makes a command of its own, such as roles in
  osapi token roles list, has a SubDoc: the one line of help of that command.
  Fix: Add command.SubDoc("segment", "...") to Define, with what the segment's
  commands do, such as Read the roles and what each grants.
- `handler`: The group Uses a type, and each method has a handler,
  func(context.Context, D, In) (Out, error), with D that type and In and Out the
  method's types. Fix: Add command.Uses\[Deps]() to Define, and .Handle(fn) to
  each method, with the signature that the message names.
- `tag`: Each tag key of an output field is one of command.KnownTags. Each
  render tag is one the output knows: -, bytes, age, time, duration,
  duration=ms, status or short=N with N above 0. An empty tag is on a list, and
  a title tag is on the blank field _ of the output type. Fix: Use a render
  value that osapi cli command spec get lists, put empty on a list field, and
  put title on a field _ struct&#123;&#125; of the output type.
- `says`: A method that answers with no body says what occurred: its name has a
  past tense, such as admitted for admit, or .Says gives the text. Fix: Add
  .Says("&lt;state&gt;") to the method, such as .Says("draining") for drain.
- `prints`: Each method that prints a body, and is not a process, has an Example
  that Prints gives its output: a value of the method's output type, which its
  docs page shows as the CLI prints it. Each process has an Example that Logs
  gives the lines it writes when it starts, which its docs page shows as its log
  handler writes them. Prints and Logs follow the Example that they show. Fix:
  Add .Prints(sample) after the Example that it shows, with sample a value of
  the method's output type, in samples.go of the group. For a process, add
  .Logs(lines...) after its Example, with the lines in samples.go of the group.

```bash
osapi cli command validate
```

## Examples

```bash
osapi cli command validate
osapi cli command validate --json
```

## Result

The command prints:

```ansi
$ osapi cli command validate
[1;38;2;0;213;255mValidation:[0m
  [38;2;10;143;176mLocal groups:[0m  17
  [38;2;10;143;176mGroups:[0m        57
  [38;2;10;143;176mValid:[0m         yes
```

With `--json`, the command prints:

```json
{
  "local": 17,
  "groups": 57,
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
