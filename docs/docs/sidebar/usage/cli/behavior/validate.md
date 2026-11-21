# Validate

Check every behavior against the rules the platform relies on.

It prints each problem with its behavior, method, rule and message, then the fix
of each rule, and exits 1 when a rule fails. `osapi behavior spec get` describes
each part of a behavior.

The rules:

- `name`: The behavior, its group and each method have a name of lowercase
  letters and digits, starting with a letter: the generators make Go names of
  it, so a dash does not compile. No behavior or group is named all. Fix: Rename
  it, such as sysctl or updatepackages. A name that is a Go keyword, such as
  package, takes osapi behavior create --package for its Go package.
- `placement`: The behavior has a placement. Fix: Add behavior.Node to Define.
- `doc`: The behavior and each method have a description. No description or
  .Explain text has &lt; &gt; &#123; or &#125;: the docs site reads them as
  markup, and the page fails to build. Fix: Give Define a one-line description,
  and each method .Describe("..."). Name a placeholder with an example, such as
  osapi-dns-eth0.yaml for eth0, not osapi-dns-&lt;interface&gt;.yaml.
- `methods`: The behavior has one or more methods. Fix: Add behavior.Read or
  behavior.Write to Define.
- `duplicate`: Each behavior is registered one time, and each method name is
  used one time. Fix: Rename or remove the second one.
- `verb`: A read is named list, get or status. A write is named create, update
  or delete. Fix: Rename the method to the standard verb that the message names.
- `route`: No two methods have the same route. Fix: Give one of them a key with
  .Key("field").
- `key`: The key of a method is a field of its input, and is required: it is a
  path segment, which is never empty. Fix: Give .Key the JSON name of an input
  field, or add the field to the input, and tag it validate:"required,...".
- `schema`: Each input is a struct, and each input and output type has a JSON
  Schema. A struct that an input or output holds is named in one spec for every
  behavior, so no two behaviors have a different type of the same name. Fix: Use
  structs of strings, numbers, bools, lists and structs. Declare a custom
  validate tag with behavior.Validator. Rename a held struct that another
  behavior's type already names, such as Usage to MemoryUsage, or hold that
  behavior's type.
- `output`: Each output is a struct or a list of structs, with no field named
  hostname, status or error. Each field has a json name, no field is an array,
  and each render tag is one the output knows (osapi cli command spec get lists
  them). Fix: Put the value in a struct, rename a field that has one of those
  names, tag each field json:"name", use a slice for an array, and use a render
  value the output knows.
- `changed`: The output of a write has Changed bool with json:"changed". Fix:
  Add Changed bool `json:"changed"` to the output, and set it to true only when
  the call changed the host.
- `span`: A write records on its span the item that it acts on, so that a trace
  shows what each call changed. Fix: Name the item with .Key("field"). A method
  without a key, such as a create, tags that field telemetry:"attr".
- `template`: A field tagged template:"&lt;vars&gt;" is a string, and
  &lt;vars&gt; is the JSON name of an input field of type map[string]string.
  Fix: Tag the string field that holds the template with template:"vars", and
  add Vars map[string]string `json:"vars"` to the input.
- `validate`: Each input field has a validate tag: its rules, or omitempty for
  an optional field with none. The spec carries the whole tag, and the server
  answers 400 to a value that breaks it. Fix: Add validate:"..." to the field,
  such as validate:"required,min=1,max=255", or validate:"omitempty" for an
  optional field with no other rule.
- `optional`: An optional field of a write's request body, with omitempty in its
  validate tag, has omitempty in its json tag too, so the client leaves it out
  when it is empty. The endpoint DSL owns the rule; controller endpoint validate
  checks it the same way. Fix: Add ,omitempty to the field's json tag, such as
  json:"version,omitempty".
- `input`: Each input field has a json name and a doc, has no default, which the
  server does not apply, is not an array, and is not named hostname or target:
  every request has the target host as its hostname path segment and its
  --target flag. Fix: Tag the field json:"name" and doc:"...", drop the default
  tag and make the provider treat the empty value as the default, use a slice
  for an array, and rename a field called hostname or target.
- `fails`: Each rule of an input field has a .Fails input that breaks it, which
  the API refuses with 400. An endpoint's required needs one, with the field
  empty; a behavior's needs none, as its generated tests leave the field out. A
  path field's required, and a min of 1 on a string that is required, optional
  or in the path, need none: their only bad value is empty. Fix: Add
  .Fails("&lt;what is wrong&gt;", &lt;Input&gt;&#123;...&#125;) to the method,
  with the other fields valid and the field's value breaking the rule that the
  message names, such as a name of 256 characters for max=255.
- `secret`: No input field is both secret:"true" and telemetry:"attr". Fix:
  Remove telemetry:"attr" from the secret field. A secret is never recorded on a
  span.
- `surface`: A behavior with CLI also has API. An internal behavior has no API
  or CLI. Each input field of a CLI behavior has a type a flag holds: string,
  bool, int, []string, map[string]string or []byte (a file). The CLI owns the
  list; its generator builds each. Fix: Add behavior.API, remove API and CLI
  from the internal behavior, or give the field a type that a flag can hold.
- `example`: Each method of a CLI behavior has an example. Each example and
  refusal has a description, and the input and output types of its method. A
  refusal has error text and no effects. Fix: Add .Example("...",
  In&#123;...&#125;, Out&#123;...&#125;) to the method, and correct the example
  or refusal that the message names.
- `settings`: Settings are a struct. Each field has a mapstructure name and a
  doc. Fix: Add mapstructure:"name" and doc:"..." to each settings field.
- `platform`: The behavior has a provider for one or more platforms, each
  platform one time, and each provider implements the provider type of each
  method. Fix: Add behavior.Platform("debian", newDebian), remove the second
  declaration, or add the missing method to the provider.
- `requires`: Each required behavior is registered, and no behavior requires
  itself through others. The chain layer owns this rule: osapi behavior chain
  validate. Fix: Import the required behavior in internal/behaviors/all, or
  remove the requirement that makes the cycle.
- `internal`: Another behavior requires each internal behavior. Nothing else can
  call it, because it has no API and no CLI. The chain layer owns this rule:
  osapi behavior chain validate. Fix: Add behavior.Requires(it) to the behavior
  that uses it, or remove behavior.Internal if an operator calls it.
- `writes`: Each path given to Writes is "_" or an absolute path.Match pattern.
  Fix: Use a pattern such as "/etc/sysctl.d/osapi-_.conf".

`go test` runs the examples and refusals of each behavior on a fake host, and
checks these. Each failure ends with its fix:

- `platforms`: Each platform that a method serves runs at least one of its
  examples. Fix: Add an example for the platform, or name it in an example's
  OnPlatform.
- `output`: Each example runs without an error, and returns the output that it
  declares. Fix: Correct the output of the example, or the provider that returns
  a different one.
- `effects`: An example of a write that changes the host declares its effects,
  and the host has each one after the call. Fix: Add Wrote, Removed or Ran to
  the example, or correct the effect that the provider did not make.
- `idempotent`: A second call with the same input reports changed: false. Fix:
  Compare the host with the input before the provider acts, and change nothing
  when they agree.
- `errors`: The provider does not panic when a file, command or store call
  fails. Fix: Check each error, and return it. Do not use the zero value of a
  call that failed.
- `state`: The provider does not panic when a record in the file-state bucket is
  damaged. Fix: Check the record before the provider uses it, and return an
  error when it is not complete.
- `refusal`: Each refusal declared with .Fails fails, with an error that
  contains its text. Fix: Refuse the input in the provider or with a validate
  tag, or correct the error text of the refusal.
- `traversal`: No string input writes a file outside the directories that the
  example writes. Fix: Add a validate pattern to the field, or check it in the
  provider and declare the refusal with .Fails.
- `newline`: No string input adds a line to a file. Fix: Refuse and in the
  field, with the validate tag no_linebreak.
- `option`: No string input reaches a command as an option. Fix: Refuse a
  leading - in the field, or pass -- before it.
- `shell`: No input reaches a shell script. Fix: Run the program with its
  arguments, not through sh -c.
- `runs`: Each program that the provider runs is declared. Fix: Add
  behavior.Runs("program"), or behavior.Sudo("program") for a program run as
  root.
- `writes`: Each path that the provider writes, removes or changes the mode of
  is declared, also through a required behavior. Fix: Add a pattern that covers
  the path to behavior.Writes.
- `file-span`: Each change to a file goes through the atomicfile package, which
  writes it atomically and makes its file.write, file.remove or file.chmod span.
  Fix: For a file that osapi manages, require managedfile and call it with
  behavior.Call. Otherwise use atomicfile.Write, atomicfile.Remove or
  atomicfile.Chmod with the ctx of the method, not h.FS.
- `keeps-state`: A behavior that uses the file-state bucket declares it. Fix:
  Add behavior.KeepsState, or require the behavior that keeps the record.
- `call`: A provider calls the methods of a required behavior through
  behavior.Call, so that the trace shows each call as a step. Fix: Call it with
  behavior.Call(ctx, other.Behavior, other.Method, provider, in).

```bash
osapi behavior validate
```

## Examples

```bash
osapi behavior validate
osapi behavior validate --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
