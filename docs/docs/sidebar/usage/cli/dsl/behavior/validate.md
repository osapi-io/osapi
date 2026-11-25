# Validate

Check every behavior against the rules the platform relies on.

It prints each problem with its behavior, method, rule and message, then the fix
of each rule, and exits 1 when a rule fails. `osapi dsl behavior spec get`
describes each part of a behavior.

The rules:

- `name`: The behavior and each method have a name of lowercase letters and
  digits, starting with a letter: the generators make Go names of it, so a dash
  does not compile. No behavior is named all. Fix: Rename it, such as sysctl or
  connectivity. A name that is a Go keyword, such as package, takes osapi dsl
  behavior create --package for its Go package.
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
  or delete. An action, declared with Action, is named for what it does with one
  of the CLI's actions, such as restart, and not with a standard verb. Fix:
  Rename the method to the standard verb that the message names, or declare a
  write named for what it does with behavior.Action and a name from osapi dsl
  command spec get.
- `route`: No two methods have the same route. Fix: Give one of them a key with
  .Key("field").
- `key`: The key of a method is a field of its input, and is required: it is a
  path segment, which is never empty. Its name in the spec, the TypeName and its
  title, is not also the name of a row type of the behavior. Fix: Give .Key the
  JSON name of an input field, or add the field to the input, and tag it
  validate:"required,...". When the message names a type, rename the type, such
  as Part to PartEntry.
- `parent`: A child is under a registered behavior. A child declared with Of is
  under an item of the parent, so the parent's items have a key, such as name
  for a user. A child declared with Under is under the parent's collection, and
  its name is not the name of an action of the parent, whose route it would
  take. Fix: Give Of a behavior that has a keyed method, such as user, or use
  Under for a parent whose items have no key, such as log. Rename a child of
  Under that has the name of an action of the parent. The child takes the
  parent's path, command and permissions.
- `schema`: Each input is a struct, and each input and output type has a JSON
  Schema. A struct that an input or output holds is named in one spec for every
  behavior, so no two behaviors have a different type of the same name. Fix: Use
  structs of strings, numbers, bools, lists and structs. Declare a custom
  validate tag with behavior.Validator. Rename a held struct that another
  behavior's type already names, such as Usage to MemoryUsage, or hold that
  behavior's type.
- `output`: Each output is a struct or a list of structs, with no field named
  hostname, status or error. Each field has a json name, no field is an array,
  and each render tag is one the output knows (osapi dsl command spec get lists
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
  &lt;vars&gt; is the JSON name of an input field of type map\[string]string.
  Fix: Tag the string field that holds the template with template:"vars", and
  add Vars map\[string]string `json:"vars"` to the input.
- `validate`: Each input field has a validate tag: its rules, or omitempty for
  an optional field with none. The spec carries the whole tag, and the server
  answers 400 to a value that breaks it. Fix: Add validate:"..." to the field,
  such as validate:"required,min=1,max=255", or validate:"omitempty" for an
  optional field with no other rule.
- `optional`: An optional field of a write's request body, with omitempty in its
  validate tag, has omitempty in its json tag too, so the client leaves it out
  when it is empty. The endpoint DSL owns the rule; dsl endpoint validate checks
  it the same way. Fix: Add ,omitempty to the field's json tag, such as
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
- `secret`: No secret.Value input field is tagged telemetry:"attr", is the key
  of a method that the API serves, or is an input of a read that the API serves:
  the key goes in the path and a read's input in the query, and a URL goes in
  logs. Fix: Remove telemetry:"attr" from the secret field, give the method a
  key that is not a secret, or make the method a write, which sends its input in
  the body.
- `before`: A .Before step is func(context.Context, In) (In, error), on a method
  that the API serves and that has a secret.Value input field. Each input field
  tagged before:"set" has a step to set it. Fix: Give the step the method's
  input type in and out, serve the behavior, give the plain secret the type
  secret.Value, and tag only the fields that the step fills before:"set".
- `tag`: Each struct tag of an input field has a known key (command.KnownTags),
  and telemetry is "attr". A secret is the type secret.Value: the tag secret is
  not a key. Fix: Correct the key or the value: a misspelled telemetry tag
  leaves the field off the span. Give a secret the type secret.Value of
  internal/secret in place of the tag secret:"true".
- `surface`: A behavior with CLI also has API, and a behavior with API has an
  Area of the UI. An internal behavior has no API or CLI. Each input field of a
  CLI behavior has a type a flag holds: string, bool, int, \[]string,
  map\[string]string, \[]byte (a file), secret.Value or
  map\[string]secret.Value. The CLI owns the list; its generator builds each.
  Fix: Add behavior.API, add behavior.Area(...) to a served behavior, remove API
  and CLI from the internal behavior, or give the field a type that a flag can
  hold.
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
  itself through others. The chain layer owns this rule: osapi dsl chain
  validate. Fix: Import the required behavior in internal/behaviors/all, or
  remove the requirement that makes the cycle.
- `internal`: Another behavior requires each internal behavior. Nothing else can
  call it, because it has no API and no CLI. The chain layer owns this rule:
  osapi dsl chain validate. Fix: Add behavior.Requires(it) to the behavior that
  uses it, or remove behavior.Internal if an operator calls it.
- `writes`: Each path given to Writes is "\*" or an absolute path.Match pattern.
  Fix: Use a pattern such as "/etc/sysctl.d/osapi-\*.conf".

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
- `gohai`: A read runs each command through gohai: it reads the facts of the
  host with a gohai collector and h.Sources(), not with h.Exec. Fix: Read the
  fact with the gohai collector that has it, such as
  services.NewWith(h.Sources()). A fact that gohai does not read is a change to
  gohai first. A read of what is not a fact of the host, such as the updates of
  the package index, declares why with .NotAFact("...").
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

`go test` also checks the chain rules that need a run of the examples.
`osapi dsl chain validate` checks them too:

- `called`: A provider calls only the behaviors that its behavior requires. The
  agent builds the provider of each required behavior, and no other. Fix: Add
  the behavior that it calls to behavior.Requires, and read its provider with
  host.Require.
- `used`: An example calls each behavior that a behavior requires. A requirement
  that no example calls is not tested, or is not needed. Fix: Add an example
  whose provider calls the required behavior, or remove the requirement.
- `secret`: A call gives a secret of its input only to a secret.Value field of
  the behavior that it calls. A field of another type, such as a string, records
  the value in its chain record, its span or an audit entry. The examples run
  each chain, so their tests check it. Fix: Give the field of the called
  behavior the type secret.Value, and read it there with Reveal.

```bash
osapi dsl behavior validate
```

## Examples

```bash
osapi dsl behavior validate
osapi dsl behavior validate --json
```

## Result

The command prints:

```ansi
$ osapi dsl behavior validate
[1;38;2;0;213;255mValidation:[0m
  [38;2;10;143;176mBehaviors:[0m  33
  [38;2;10;143;176mValid:[0m      yes
```

With `--json`, the command prints:

```json
{
  "behaviors": 33,
  "valid": true
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
