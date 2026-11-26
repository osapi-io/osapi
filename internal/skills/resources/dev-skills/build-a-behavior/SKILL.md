---
name: build-a-behavior
description: Build or change an osapi dsl behavior, the typed definition under internal/behaviors/<name>/ that osapi's API routes, agent dispatch, CLI commands, OpenAPI spec, docs pages, settings, permissions and tests are generated from. Use when asked to add a behavior, manage something new on a host, add a method or a setting to a behavior, give a behavior a provider for another platform, or when a behavior fails osapi dsl behavior validate, behavior-gen-check or its tests. Also use to decide whether something should be a behavior at all, whether it should be internal or served, which platforms it runs on, whether it should require another behavior or be split into an internal one, and whether it belongs in osapi-orchestrator instead. Not for operations that run on the controller itself, such as the agents, the jobs and the audit log, which the build-an-endpoint skill covers.
---

# Build a behavior

A behavior is one directory, `internal/behaviors/<name>/`, plus one import in
`internal/behaviors/all`. Everything else is generated or inferred from its
definition.

This skill gives a procedure and decisions. It is not a catalog. The platform
tells about itself:

- `dsl behavior spec get` tells each part of the DSL, what it does and when to
  use it, the generated paths and what the platform cannot do yet.
- `dsl behavior validate --help` tells each rule and each test check, with its
  fix. Each failure also ends with its fix.
- `dsl behavior list` and `dsl behavior get` tell what each behavior accepts and
  does. This includes the building blocks.

A behavior is work on a host. An operation that runs on the controller itself,
such as the agents, the jobs or the audit log, is an endpoint. The
build-an-endpoint skill covers it:
`go run . skill get --name build-an-endpoint`.

Ask these tools. Do not use your memory. If a tool and this skill do not agree,
the tool is correct. If the platform does not have what you need, stop and say
what is missing. That is a platform change, and a person decides it. Do not add
code around it.

## Before anything

The binary embeds the UI, so `go run .` fails until it is built once:
`mise exec -- just build`. After that, `go run .` works for every command below.
Run tools through `mise exec --`, as the repository's AGENTS.md says.

## Find what you need

Ask the code and the binary before you read prose. Each one answers one
question:

| Question                                                                                                    | Ask                                                                                                                    |
| ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| What options, methods, example parts and tags exist, and when to use each                                   | `go run . dsl behavior spec get`                                                                                       |
| What rules and test checks a behavior must pass, and the fix of each                                        | `go run . dsl behavior validate --help`                                                                                |
| What a provider builder gets from the agent                                                                 | `go doc ./internal/behavior/host Host`                                                                                 |
| Which facts of the host gohai reads, how a provider builds each collector, and what an example must give it | `go run . dsl behavior spec get`, under "gohai collectors"                                                             |
| What behaviors exist already                                                                                | `go run . dsl behavior list`                                                                                           |
| Which building blocks do generic work, such as a managed file, and when to make a new one                   | `go run . dsl chain spec get`, under "Building on other behaviors" and "Building blocks"                               |
| What a behavior that requires others sends to them, and gets back                                           | `go run . dsl chain get --name <name>`: each example's calls, with their inputs and outputs; `--json` has them in full |
| What one accepts and does: routes, fields, platforms, sudo, what it uses                                    | `go run . dsl behavior get --name <name> --json`, building blocks included                                             |
| Whether a definition breaks a rule                                                                          | `go run . dsl behavior validate`                                                                                       |
| A complete example of every part                                                                            | `internal/behaviors/sysctl`                                                                                            |
| How tests fake each host service, and which pattern to copy for each kind of state                          | [references/services.md](references/services.md)                                                                       |
| The walkthrough, with a worked behavior                                                                     | `docs/docs/sidebar/development/your-first-behavior.md`                                                                 |

## Types, not strings

A value of a closed set has a type, never a string:

- **The type of `pkg/sdk/types`**, when it has one.
  `go run . dsl behavior spec get` lists each under "Types for values", and
  `go run . sdk get --name <Type>` shows its values. The permissions
  (`types.Perm...`), the roles (`types.Role...`), the verbs (`types.Verb...`)
  and the targets (`types.TargetAll`, `types.TargetAny`) are constants there
  too.
- **A named type of the package**, in its `types.go`, with `Values()` and
  `Valid()`, for a closed set that only this package has.
- The field takes the rule `validate:"enum"`. `validate` refuses `oneof`.

Use the constant, never its string: `types.ContentRaw`, not `"raw"`.
`internal/dslcheck` fails on a string literal where a typed value goes, and on a
string in the API of a DSL that `internal/dslcheck/testdata/strings.txt` does
not list with its reason.

## Decide first

Decide these with the user before you write code. Each answer becomes a
declaration. Thus an incorrect guess becomes an incorrect API, and other code
then uses that API. Ask only what the request does not tell you. Give your
recommendation with each question. Say which declaration the answer becomes.

| Ask                                                                                                                                                                                                        | Why it matters                                                                                                                                                                        | Answer becomes                                                                                                                                                                                                                                                                                                                                                      |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Does the work occur on a host, or on the controller, such as reading the registry or the job queue?                                                                                                        | A behavior changes or reads a host; the controller's own operations are endpoints                                                                                                     | This skill, or the build-an-endpoint skill                                                                                                                                                                                                                                                                                                                          |
| Is there one close already? Check `go run . dsl behavior list --json`.                                                                                                                                     | Two behaviors managing the same thing disagree about its state                                                                                                                        | Extend that one: a method, a field or a platform                                                                                                                                                                                                                                                                                                                    |
| Who calls it: operators, or only other behaviors? Would an operator ever call it from the CLI or the API? Is it only a piece of work other behaviors share, such as writing a file or reloading a service? | An operator needs routes, commands and permissions; a building block must not have them. `dsl behavior validate` fails an internal behavior that no behavior requires                 | `behavior.API` and `behavior.CLI` when an operator calls it, or `behavior.Internal` when only other behaviors do                                                                                                                                                                                                                                                    |
| Which hosts does it run on?                                                                                                                                                                                | Every other platform reports `skipped`, never fails                                                                                                                                   | `behavior.Platform("debian", ...)` per distribution, `"linux"` for every Linux host without its own, `"darwin"` for macOS                                                                                                                                                                                                                                           |
| Does one method run on fewer hosts than the others?                                                                                                                                                        | That method reports `skipped` elsewhere, and its examples run only there                                                                                                              | `.On(behavior.PlatformDebian)` on the method, and `.Sudo` or `.Runs` on the method for a program only it runs. `hostname` is the example                                                                                                                                                                                                                            |
| Does part of the work exist as an internal behavior? `dsl behavior list` marks them `(internal)`.                                                                                                          | Copying it means two copies of the hard part                                                                                                                                          | `dsl behavior create --requires <that>` writes `behavior.Requires(that.Behavior)`, the `host.Require` in the builder and a provider field. You write the calls: call each of its methods with `behavior.Call(ctx, that.Behavior, that.Method, provider, in)`, so the trace shows the call as its own step and its input is validated. A direct call fails the tests |
| Would a second behavior need part of what this one does?                                                                                                                                                   | The rule of two: the second need is when shared work becomes its own behavior                                                                                                         | Build that part as an internal behavior first, then require it                                                                                                                                                                                                                                                                                                      |
| Does a value change per call, or stay the same on a host?                                                                                                                                                  | A setting is agent config; an input is part of each request                                                                                                                           | An input field, or a field in `settings.go`                                                                                                                                                                                                                                                                                                                         |
| Which input names the thing a call acts on, such as a key or a path?                                                                                                                                       | A trace shows it, so a reader sees what each call changed. `dsl behavior validate` fails a write that records nothing                                                                 | `.Key("key")`, or `telemetry:"attr"` on a method without a key                                                                                                                                                                                                                                                                                                      |
| Which programs does it run, which as root, and does it take a secret?                                                                                                                                      | The agent checks each is installed and sudo allows it; secrets are kept out of the audit log and traces                                                                               | `behavior.Runs("program")`, or `behavior.Sudo("program")` for root; the type `secret.Value` on the field, or `map[string]secret.Value`, and `Reveal` in the provider                                                                                                                                                                                                |
| Which files does it write, remove or change the mode of, and does it keep a record in the file-state bucket?                                                                                               | Each one is a span in a trace, and the examples fail on a path or a record that is not declared. Declare the paths even when a required behavior, such as managedfile, does the write | `behavior.Writes("/etc/x.d/osapi-*.conf")`, and `behavior.KeepsState` for a record                                                                                                                                                                                                                                                                                  |
| Is it one host, one step?                                                                                                                                                                                  | A behavior acts on one host inside one job                                                                                                                                            | If it spans hosts or orders several steps, it is an osapi-orchestrator workflow, not a behavior                                                                                                                                                                                                                                                                     |
| Does it need anything that `dsl behavior spec get` lists under "Not supported yet"?                                                                                                                        | The platform cannot express it, and wiring around that breaks the generated layers                                                                                                    | Stop and tell the user what is missing                                                                                                                                                                                                                                                                                                                              |

Then decide how it operates. This decides the provider and its examples. Read
[references/services.md](references/services.md) first. Then select the pattern
that the answers identify:

| Ask                                                                                                                        | Answer becomes                                                                                                                                                                         |
| -------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Where is the truth about what it manages: a file osapi writes, a file something else owns, a program's output, or `/proc`? | Which pattern on that page the provider follows, and which host options its examples use                                                                                               |
| Does a read show the host as it is, or only what osapi wrote?                                                              | The host as it is, the way Ansible, Salt and Puppet read: every item, with whether osapi manages it as a field. A read limited to osapi's own records hides the host from the operator |
| What does "already done" look like, so a second run changes nothing?                                                       | The comparison the provider makes before acting, and the example's `changed: false` second run                                                                                         |
| Which commands does it run, and which need root?                                                                           | `h.Exec` calls; `behavior.Runs` for each program, `behavior.Sudo` instead for a privileged one                                                                                         |
| Which inputs reach a path, a line in a file or a program's arguments?                                                      | Their `validate` patterns and the refusals declared with `.Fails`                                                                                                                      |
| What should it refuse, and what does it answer when the thing is not there?                                                | Refusals, and `behavior.ErrNotFound` for a 404                                                                                                                                         |

Tests do not run real programs. The filesystem of the fake host is in memory.
Its commands give the output that the examples declare. Thus you can build and
test a behavior for Debian on any machine. Do not skip an example because the
program is not installed on your machine.

The request can leave these open: what the behavior manages, where its truth is,
or which methods it has. Then ask before you write code, because a guess becomes
an API. Use the defaults below only when the user says that they have no
preference. Do not use them in place of a question. The defaults are:

- a new behavior only if no close behavior exists
- `API` and `CLI`
- `debian`
- inputs, not settings
- no `Requires`, unless an internal behavior already does part of the work

Name each method for the command it becomes: each command under `osapi client`
is `<resource> <verb>`. The verb is `list`, `get`, `create`, `update` or
`delete`, or an action that a CRUD verb does not say, such as `restart`. A noun
is never the verb. A required input is `validate:"required"`; `--help` marks it,
so its `doc` does not say so. `TestCommandShape` in `cmd/help_style_test.go`
fails on a command that breaks either rule.

Each input field has a `validate` tag, such as
`validate:"required,min=1,max=255"`. A field with no limit is
`validate:"omitempty"`. The server answers 400 to a value that breaks a rule,
and the spec carries the rule and the whole tag. `validate` refuses an input
field with no tag. A rule of the behavior's own, such as `sysctl_key`, is a
`behavior.Validator` in the behavior's directory, with its check, its schema and
its message. A shared rule, such as `resource_name`, has its schema from
`rule.Register`, beside its validator in `internal/validation`. Without a schema
the spec does not build. The endpoint DSL and the behavior DSL use the same
rules; `spec` prints them under `validate:"..."`.

Name and shape each field so the generated layers can build it.
`dsl behavior validate` fails each of these, and its message ends with the fix:

- **The name of the behavior** is lowercase letters and digits, with no dash:
  the generators make Go names of it. It is the route, the command and the
  permission. A name that is a Go keyword, such as `package`, keeps its name and
  takes `osapi dsl behavior create --package <go-package>` for the directory.
- **Each input field** has a `json` name and a `doc`, and no `default`, which
  the server does not apply. Make the provider treat the empty value as the
  default, and say in the `doc` what it is. No input field is named `hostname`
  or `target`: every request has the target host as both.
- **An optional field of a write's body**, `validate:"omitempty,..."`, is
  `json:"name,omitempty"` too. Without it the client sends `""`, and the field's
  other rules refuse it. The endpoint layer owns this rule.
- **A key** (`.Key`) is `validate:"required"`: it is a path segment.
- **A CLI input field** is a type a flag holds: `string`, `bool`, `int`,
  `[]string`, `map[string]string` or `[]byte` (a file). The CLI owns the list.
  An `int64` or a `float64` does not compile into a flag.
- **An output field** has a `json` name and a `render` tag that the output
  knows. No field is an array; use a slice.
- **A struct that an output holds**, such as `Usage` in memory's output, is
  named in one spec for every behavior. Give it a name no other behavior's type
  has, or hold that behavior's own type, as `status` holds memory's.

The API, the agent and `behavior.Call` check these rules before each method
runs, for an internal behavior too. Do not write the check in the provider's
method; declare the rule.

Give each rule of each input field a `.Fails` input that breaks it: the method's
input, with the other fields valid and one value wrong, such as a name of 256
characters for `max=255`. The API refuses each one with 400, and a test sends
each one through the routes. `validate` names a rule that has none. In an
endpoint, a required field needs one too, with the field empty. In a behavior it
needs none, because the generated tests leave the field out. You write the bad
value, because a generator cannot know one for a custom rule.

The behavior's tests run each refusal as the agent does: the `validate` rules
first, then the method. A refusal that a rule refuses holds with the rule's
text, such as `"'max' tag"`, and the provider need not check it again. A refusal
that no tag can say, such as "not found", or a check the provider makes as well,
such as of a program's argument, holds with the provider's error.

Work that two behaviors share is a building block, not a copy. The `netplan`
behavior writes a file under `/etc/netplan`, validates it with
`netplan generate`, and applies it. When `netplan generate` refuses the file,
the file gets back the content that it had, or a new file is removed. dns,
interface and route require it. Its `get` gives the netplan ID of the definition
that matches an interface, such as cloud-init's `all` for `en*`, and each
interface that the definition matches. netplan ignores a second definition of
the same interface: dns writes under that ID, and interface refuses an interface
that a match defines. `netplan status` lists only the interfaces that netplan
manages, so a behavior asks gohai whether the host has an interface.

A provider reads the host through gohai, never directly. Find the fact under
"gohai collectors" in `go run . dsl behavior spec get`. Build that collector in
the builder with the constructor that the spec gives, such as
`dns.NewWith(h.Sources())`, so the provider reads the host that the agent or the
test gives. Each example's fake host gives what the collector reads: an `Exists`
for each file and a `Prints` for each command that the spec lists for the
platform. A fact that no collector has is a change to gohai first, by its
`docs/adding-a-collector.md`. Do not read the file or run the command in osapi
to work around a missing fact.

The behaviors that a behavior calls, and the ones that they call, are its chain.
`go run . dsl chain list` shows each chain, and `go run . dsl chain validate`
checks it: each requirement is called by an example, and a provider calls only
the behaviors that it requires.

A write example runs on a live host in the integration tests. Thus it uses an
item that it cannot harm, such as `osapi0`, a dummy interface that the
integration job makes. An effect that is not safe on a live host, such as a
default route, goes in a unit test of the provider, not in an example.

Give each write example a `behavior.Reads` with what a read answers after it,
such as
`behavior.Reads(GetInput{Key: "vm.swappiness"}, Entry{Key: "vm.swappiness", Value: "10", Managed: true})`.
The integration test runs that read on the live host after the example, so a
write that reports a change and makes none fails. With no field names, it
compares each field of the want that is not zero; name a field to compare a zero
value, such as `managed` for false. A list matches a list that has each of its
values, so an interface's `addresses` can have its link-local address too. After
a delete, `behavior.ReadsNone` says that no row has the want, or that the read
reports not found. A write whose result no read shows, such as a command's run,
gets `.ReadsNothing(why)` on the method. `dsl behavior validate` fails a live
write example with none of them.

## Build a new one

1. Write it:
   `go run . dsl behavior create --name <name> --description "<one line>"`, with
   `--package <go-package>` when the name is a Go keyword, such as
   `--name package --package packages` (the directory is then that package), and
   `--requires <behavior>[,<behavior>...]` for each behavior that it is built
   on, such as `--requires managedfile`, then
   `go generate ./internal/behaviors/<name>/mocks ./internal/behaviors/all`. The
   behavior is directly under node, by its own name. To nest it under a parent
   behavior, add `behavior.Of(parent.Behavior)` to put it under an item of the
   parent, as the SSH keys are under a user, or
   `behavior.Under(parent.Behavior)` to put it under the collection, as the
   sources are under the log. Never nest a behavior only to sort it: its area
   sorts it. The result builds, validates and passes every test before you
   change it. With `--requires`, one test fails: the examples fail the rule
   `used` until a provider calls each required behavior with `behavior.Call`.
   The comment on the provider field in `<platform>.go` has the call to write.
   Then give the mock of each required provider its expectations in
   `<platform>_public_test.go`.
2. Change it in this order, because each depends on the one before:
   `settings.go`, `types.go`, `provider.go` and `<platform>.go`, `methods.go`,
   then `behavior.go`. Use the host only through `host.Host`. Then the fake host
   can test it. Compare before you act. Then a second run reports
   `changed: false`. A building block's own fields, such as what `managedfile`
   accepts, come from `dsl behavior get` on it.
3. Give each method a useful example. The example names the host that it runs on
   and the changes on that host. Declare what the method refuses with `.Fails`.
   `dsl behavior spec get` lists the parts of an example. The examples run the
   provider on a fake host and become the generated HTTP and CLI tests.
4. Keep a test for every file with code, named for it: `methods_public_test.go`
   on the mocked provider, and `<platform>_public_test.go` for what an example
   cannot show. After you change `provider.go`, run
   `go generate ./internal/behaviors/<name>/mocks`.
5. Run the checks below until they pass.

## Design docs

AGENTS.md puts the design in `osapi-io/specs` before you build. For a behavior,
the definition is the design. The generated pages and `dsl behavior get` publish
its methods, types, settings, examples and refusals. The page of the behavior
has its platforms, settings and the behaviors it uses. The CLI page of each
method has its examples, what each example does on the host, its refusals and
its permission. Thus a new behavior does not need its own specs page. Change
`components/osapi/behaviors.md` in specs only when the work changes the platform
itself: a new option, rule, surface or check that every behavior gets.

## Check it

```bash
go generate ./internal/behaviors/all
mise exec -- just generate   # joins its spec into the main one: SDK, UI, API docs and CLI pages
go test ./internal/apicontract/   # the contract of the API; -update when the change is on purpose
go run . dsl behavior validate
go test ./internal/behaviors/<name>/... ./internal/behaviors/all/
mise exec -- just test
```

CI also runs the write examples against a real agent on Linux, as root, in its
`Integration` step, with `OSAPI_INTEGRATION_WRITE_BEHAVIORS=1`. So a write
example sets its starting state with `After`, not with `Prints` or `Exists`
alone: a live host has its own files and programs, and only a call can put it in
the state the example needs. A behavior with no `delete` gets its item back
after each example, through `get` and `update`.

A failure of `dsl behavior validate` or of a behavior's tests ends with its fix.
`go run . dsl behavior validate --help` lists each rule and each test check
before you break one. The failures below come from elsewhere:

| Failure names                                     | It means                                                                                                                                                                                                                                        |
| ------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `<name> runs <program>, which this host lacks`    | CI only: install the program's package on the runner in `.github/workflows/go-int.yml`                                                                                                                                                          |
| `TestAPI<Method>PublicTestSuite/...`              | The API answer for that case in `gen/api_<method>_public_test.go`: the status, the body, or the spec it breaks                                                                                                                                  |
| `is not mounted` or `is not attached`             | Run `go generate ./internal/behaviors/all`; the gen package is missing                                                                                                                                                                          |
| `behavior output is stale`                        | Run `mise exec -- just generate` and keep the result                                                                                                                                                                                            |
| `TestJoinedSpec`                                  | The full spec changed, as a new served behavior or a cutover changes it. Write the contract again with `go test ./internal/apicontract -run TestContractPublicTestSuite/TestJoinedSpec -update`, and keep the diff: it is the change to the API |
| `cli-docs-gen-check` fails                        | Run `mise exec -- just cli-docs-generate` and keep the result                                                                                                                                                                                   |
| `is served by the <name> behavior and by <dir>`   | A route a hand-written domain still serves: this is a cutover, so delete that domain in the same change                                                                                                                                         |
| `is both a hand-written command and a behavior's` | The same, for a command under `osapi client node`                                                                                                                                                                                               |
| `a behavior is below 100% coverage`               | A branch no test reaches: add an example, or a case to the test named for the file                                                                                                                                                              |

Never edit a path that `dsl behavior spec get` lists under "Generated, never
edited". Change the definition and generate. Do not write CLI docs for a
behavior.

## Not yet

`dsl behavior spec get` lists what the platform cannot express yet, under "Not
supported yet". If a request needs one, say so and stop. Do not build around it.

A behavior needs no edits outside its directory. The only exception is a package
that CI's integration job says is missing from its runner. If you must edit
other code, for example `cmd/agent_setup.go`, `internal/validation` or
`internal/authtoken`, the platform does not have a declaration that you need.
Say so. Do not add the code by hand.

When the change touches what the DSLs share, such as the spec and checks tables,
typed values and `command.KnownTags`, the permissions that the role rule reads,
`genfile.Writer` or the generated DSL pages, read the build-a-dsl skill first:
`osapi skill get --name build-a-dsl`.
