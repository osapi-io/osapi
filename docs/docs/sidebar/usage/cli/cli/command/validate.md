# Validate

Check every local group against the rules of a definition.

It exits 1 when a rule fails, and names the group, the method, the rule, what is
wrong and the fix. Each problem in `--json` has `group`, `method` (when the
problem is in one), `rule`, `message` and `fix`. `osapi cli command spec`
describes each part of a group.

The rules:

- `name`: The group, each word of Under, each method and each Sub have a name of
  lowercase letters and digits, starting with a letter. Fix: Name the group for
  what its commands are about, such as token, and each method for what it does,
  such as generate.
- `methods`: The group has at least one method. Fix: Add a method with
  command.Method, with its input and output types and its name.
- `unique`: No two methods of a group have one name: the name is the key of the
  method's renderer. Fix: Give each method its own name.
- `verb`: The last word of each command is a verb (list, get, create, update,
  delete) or an action in command.Actions, as TestCommandShape holds each
  command to. Fix: Name the method for a verb, and put a noun before it with
  .Sub, such as roles list. Add a word to command.Actions only for a real
  action, never for a noun.
- `help`: Each method has a Short, and a method with input has an Example. Fix:
  Add .Short("&lt;what it does&gt;") and .Example("--flag value") to the method.
- `flag`: Each value is a flag: the input is a struct, each field has a json
  name that is its flag with _ as -, no flag or in tag, and a type a flag holds
  (string, bool, int, []string, map[string]string or []byte). A command takes no
  positional argument. Fix: Make the input a struct, tag each field json:"name",
  drop the flag and in tags, and use one of the flag types.
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
- `handler`: The group Uses a type, and each method has a handler,
  func(context.Context, D, In) (Out, error), with D that type and In and Out the
  method's types. Fix: Add command.Uses[Deps](<>) to Define, and .Handle(fn) to
  each method, with the signature that the message names.

```bash
osapi cli command validate
```

## Examples

```bash
osapi cli command validate
osapi cli command validate --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
