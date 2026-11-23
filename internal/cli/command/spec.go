// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package command

import (
	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/validation/rule"
)

// The kinds of term in Spec, in the order that a group is written.
const (
	KindDefinition = "definition"
	KindOption     = "option"
	KindMethod     = "method"
	KindRefine     = "refine"
	KindTag        = "tag"
	KindOutput     = "output"
	KindWiring     = "wiring"
	KindDeps       = "deps"
	KindCheck      = "check"
	KindGenerated  = "generated"
)

// Spec is every term of the command DSL. osapi cli command spec get prints it,
// so an author, or an agent, learns the DSL from the binary that it builds.
// A test fails on an exported option, method or refinement of this package
// without an entry.
var Spec = guide.Spec{
	Kinds: []guide.Kind{
		{Kind: KindDefinition, Title: "Definition"},
		{Kind: KindOption, Title: "Options of Define"},
		{Kind: KindMethod, Title: "Methods"},
		{Kind: KindRefine, Title: "Refinements of a method"},
		{Kind: KindTag, Title: "Tags on input fields"},
		{Kind: KindOutput, Title: "Output, from the output type"},
		{Kind: KindDeps, Title: "What a Deps takes"},
		{Kind: KindWiring, Title: "Wiring"},
		{Kind: KindCheck, Title: "What holds a group to its rules"},
		{Kind: KindGenerated, Title: "Generated, never edited"},
	},
	Terms: []guide.Term{
		{
			Kind: KindDefinition,
			Name: `Register(Define(name, doc, options...))`,
			Does: "Declares a group of local commands and adds it to the registry. A local command runs in the CLI process and calls no API, such as osapi token generate. Its command is osapi <name>, and doc is its help.",
			When: "One time, in <group>.go under internal/cli/commands/<group>: var Group = command.Register(command.Define(...)). Add its import to internal/cli/commands/all. A command that calls the API is an endpoint or a behavior, not a local command.",
		},

		{
			Kind: KindOption,
			Name: "Uses[Deps]()",
			Does: "Names the struct of what the group's methods use, such as the config, the file system or the signing of tokens. The CLI gives one with Provide when a command runs, and a test gives one with fakes.",
			When: "In each group. Keep it to interfaces, so a test can fake each one. An empty struct needs no wiring.",
		},
		{
			Kind: KindOption,
			Name: `Long("text")`,
			Does: "Is the rest of the help of the group's command, after its one line. A method's .Long is the help of its own command.",
			When: "When the group's command needs more than its one line, such as what each of its commands reads.",
		},
		{
			Kind: KindOption,
			Name: `Fail()`,
			Does: "Is the error a handler returns with its output when the output says what failed, such as the problems of a config: return out, command.Fail(). The output prints as it does on success, then the command exits 1 with no message of its own.",
			When: "For a check whose answer is the report, such as config validate, so a script or CI fails on it. For any other failure, return the error itself.",
		},
		{
			Kind: KindOption,
			Name: `Flat()`,
			Does: "Makes the group add no word of its own: each method is a command under the group's parent, the root or the words of Under. A command directly under the root has a word of command.Roots, such as version; any other has a verb or an action. The group's name stays its identity, for cli command get --name and its package.",
			When: "For a command that stands alone, such as osapi version, or that goes under a parent that holds other things, such as osapi controller start. Not for a family of commands under one noun, such as token.",
		},
		{
			Kind: KindOption,
			Name: `SubDoc("segment", "short")`,
			Does: "Is the one line of help of the command that a Sub segment makes, such as roles in osapi token roles list. validate fails a Sub segment that makes a command and has none.",
			When: "For each Sub segment of the group whose word is not the method's.",
		},
		{
			Kind: KindOption,
			Name: `Under(Parent, ...)`,
			Does: "Puts the group's command under the command of the words, such as ParentController for osapi controller endpoint. Use a Parent constant: validate refuses a word outside KnownParents, because Attach makes a word that has no command, so a misspelled word was a new command with no error. Attach uses a hand-written command that cmd/commands.go names.",
			When: "When the commands belong to a command that is there already, such as controller. Without it, the group is under the root.",
		},

		{
			Kind: KindMethod,
			Name: `Process("name")`,
			Does: "A command, osapi <group> <name>, that runs a process until the command's context ends, such as osapi nats server start. It has no input and prints no output; .Setup sets what it runs. The context is the one the root makes and cancels on SIGINT or SIGTERM.",
			When: "For a command that starts a component and runs until a signal. Its name is an action, such as start.",
		},
		{
			Kind: KindMethod,
			Name: `.Setup(fn)`,
			Does: "Is func(ctx context.Context, d D) (command.Running, error): it makes the process's components with ctx, the command's context, so a signal stops a setup that waits. Running is the components, each with Start() and Stop(ctx), and a Cleanup. command.Serve starts them, waits until ctx ends, stops them together, then runs Cleanup; the stop and the cleanup get context.WithoutCancel(ctx), which keeps its values, with command.ShutdownTimeout, 10 seconds. An error is the command's error, and it exits 1.",
			When: "On each Process, in place of .Handle. Make each component in Setup and start nothing: Serve starts them.",
		},
		{
			Kind: KindMethod,
			Name: `Then(running, fn)`,
			Does: "Is running with fn after its Cleanup, such as the stop of a tracer, which flushes what the components recorded as they stopped.",
			When: "In a Setup that wraps the components of another, such as osapi start, which runs the NATS server, the controller and the agent under one tracer.",
		},
		{
			Kind: KindMethod,
			Name: `.Flag("key", ...)`,
			Does: "Adds a flag to a process for each config key, such as agent.max_jobs: --agent-max-jobs, the key with its dots and underscores as dashes, as OSAPI_AGENT_MAX_JOBS is its environment variable. The flag has the key's type, default and doc, from the config, and sets the key over the environment and the file.",
			When: "On a Process, for each key a person sets when they start it. A Method's flags are its input, so it has no Flag.",
		},
		{
			Kind: KindMethod,
			Name: `Method[In, Out]("name")`,
			Does: "A command, osapi <group> <name>, that takes In and prints Out. Each field of In is a flag. Out prints as JSON with --json, and from its type without it, by the output style of the CLI.",
			When: "For each command of the group. Its name is the command's last word: a verb (list, get, create, update, delete) or an action in command.Actions, never a noun.",
		},

		{
			Kind: KindRefine,
			Name: ".Handle(fn)",
			Does: "Sets the function that runs the command: func(ctx context.Context, d Deps, in In) (Out, error). The generated command reads the flags into In, checks it against its validate tags, and calls it. An error from it prints, and the command exits 1.",
			When: "On each method. Validate checks the signature.",
		},
		{
			Kind: KindRefine,
			Name: `.Short("text")`,
			Does: "Sets the command's one line of help. It starts with a capital letter and has no period at the end; a test of every command holds this.",
			When: "On each method.",
		},
		{
			Kind: KindRefine,
			Name: `.Long("text")`,
			Does: "Sets the rest of the command's help, in sentences.",
			When: "When a person needs more than the one line, such as what the command reads.",
		},
		{
			Kind: KindRefine,
			Name: `.Example("--flag value", "doc")`,
			Does: "Adds an example of the command: the flags after osapi <group> <method>, and, when given, what it shows, as a # line above it. The help and the CLI page show each one.",
			When: "At least one for each method with input: the call a person makes most.",
		},
		{
			Kind: KindRefine,
			Name: `.Prints(sample)`,
			Does: "Gives the last Example its output: a value of the method's output type, which the CLI page shows as the CLI prints it, and as JSON for --json. The page of a method with no output shows its one line for each example, from the example's flags, with no Prints.",
			When: "After the Example that it shows, on each method that prints a body and is not a process: validate fails one with no Prints. Put the sample in samples.go of the group, with each time in UTC.",
		},
		{
			Kind: KindRefine,
			Name: `.Sub("segment")`,
			Does: "Puts the command under a command of its own in the group: osapi <group> <segment> <method>, such as token roles list. The method's identity is its segment and its name, so a group can have get and spec get.",
			When: "When the command is about a noun of the group: the noun is the Sub, and the method is the verb.",
		},
		{
			Kind: KindRefine,
			Name: `.Says("state")`,
			Does: "Sets what the one line of a method with no output says, after the key of the input: a state, such as draining. Without it, the line says the past tense of the method, such as admitted for admit. An endpoint method has the same refinement.",
			When: "On a method with no output whose result is a state, not the past tense of its name. validate fails a method with no output, no past tense and no Says.",
		},
		{
			Kind: KindRefine,
			Name: `.Fails("doc", in)`,
			Does: "Declares an input that the command refuses with a usage error, exit code 2: the method's input, with the other fields valid and one rule broken. TestRefusals in internal/cli/commands/all runs each one.",
			When: "For each rule of each input field, required included, as validate requires.",
		},

		{
			Kind: KindTag,
			Name: `json:"name"`,
			Does: "Names the field, its flag (with _ as -) and its key in the JSON of the output.",
			When: "On each field.",
		},
		{
			Kind: KindTag,
			Name: `doc:"..."`,
			Does: "Is the help of the field's flag. A rule of oneof adds its values in brackets.",
			When: "On each input field: a capital letter first and no period at the end.",
		},
		{
			Kind: KindTag,
			Name: `short:"r"`,
			Does: "Gives the field's flag a one-letter shorthand, such as -r.",
			When: "When the CLI uses that shorthand for the flag elsewhere.",
		},
		{
			Kind: KindTag,
			Name: `default:"value"`,
			Does: "Gives the field's flag its value when the caller gives none. A field with a default is never a required flag.",
			When: "For an option with a value that most calls use.",
		},
		{
			Kind: KindTag,
			Name: rule.TagName,
			Does: "The rules of the value, which Run checks before the handler runs: a value that breaks one is a usage error that names the flag, and the command exits 2. required makes the flag required. The rules are those of internal/validation, which endpoints and behaviors use; dive applies the rules after it to each value of a []string. A custom rule, such as permission, is declared with rule.Declare.",
			When: rule.TagWhen,
		},
		{
			Kind: KindTag,
			Name: "the type of a field",
			Does: "Is the kind of its flag: a string, a bool, an int, a []string (a flag given once for each value, or with commas), a map[string]string (key=value pairs) or a []byte (the path of a local file, whose content the field gets).",
			When: "Pick the type of the value a person types.",
		},

		{
			Kind: KindOutput,
			Name: "Render(p, group, out)",
			Does: "Prints the output of a method through a Printer, as the CLI prints it. The terminal and the pages of the docs site both use it, so a page written from a command's output shows what the command shows.",
			When: "To write an output somewhere other than the terminal, such as a page of the reference.",
		},
		{
			Kind: KindOutput,
			Name: "a list: Items []T, with TotalItems or not, or []T",
			Does: "Prints a heading with the count, such as Files (6):, then a table with one column for each field of T. A list with fewer items than TotalItems counts as (3 of 120). An empty list prints its heading with (0), then None. and the hint of its empty tag.",
			When: "For each method that returns a list. Name the fields Items and TotalItems, and add no other field.",
		},
		{
			Kind: KindOutput,
			Name: "one object",
			Does: "Prints a heading, then an aligned key-value block, one pair for each field. An object in it is a section under the block, with its label as the heading. A list in it is a table under the block, with its count. An object with one field that is an object or a list is an envelope: the value of the field prints, and the envelope does not.",
			When: "For each method that returns one thing.",
		},
		{
			Kind: KindOutput,
			Name: "no output",
			Does: "A method whose output type has no field, such as struct{}, or whose answer has no body (204), prints one line: the key of the input and what occurred, such as m-1 admitted. With --json, it prints nothing.",
			When: "For a write with nothing to return. The text is the past tense of the method, or what .Says gives.",
		},
		{
			Kind: KindOutput,
			Name: "the rules with no options",
			Does: "The columns and the keys have the order of the fields. A field with omitempty or omitzero and no value is not shown. A bool is yes or no. A float32 or float64 has two decimals, such as 0.00. A []string is its values, joined with commas. A time.Time is in the one time format of the CLI. A label is the json name in sentence case, with an initialism such as ID in capitals: machine_id is Machine ID. A column header is the label in capitals. A table leaves out a column that has no value in any row. In a table cell, an object is its values, with a space between them, such as Ubuntu 24.04. A list of objects is the first field of each, with commas, such as HighLoad, DiskPressure. The heading is the name of the group in sentence case, with s for a list.",
			When: "To change the order of the output, change the order of the fields.",
		},
		{
			Kind: KindOutput,
			Name: `render:"-|bytes|age|time|duration|duration=ms|status|short=N|body"`,
			Does: "Changes how a field prints: - leaves it out (--json has it), bytes is a size such as 3.4 GB, age is the time since a time or a number of seconds, time is a time as text in the one time format, duration is a Go duration, duration=ms is an integer of milliseconds as a duration, status is a status word in its color, short=N is the first N characters with ... after them, in a table only: a key-value block shows all of the value, and body makes the field the whole output, as it is, such as the text of a document; the other fields are only in --json.",
			When: "On an output field whose raw value a person cannot read at a glance. validate fails a value that is not one of these.",
		},
		{
			Kind: KindOutput,
			Name: `label:"..."`,
			Does: "Gives the field a different label, and the column the same text in capitals.",
			When: "When the json name does not say what the value is to a person, such as Last seen for registered_at.",
		},
		{
			Kind: KindOutput,
			Name: `none:"..."`,
			Does: "Is the text of a field with no value: absent, null, empty or a zero time. A field with omitempty and a none tag is shown, with the text. A table keeps a column with a none tag, because no cell of it is empty.",
			When: "When no value means something, such as never for a time or (deleted) for a hash.",
		},
		{
			Kind: KindOutput,
			Name: `empty:"..."`,
			Does: "On a list field: the hint that an empty list prints after None., such as Upload one with osapi client file upload.",
			When: "On each list that a person can add to, to tell what to do next.",
		},
		{
			Kind: KindOutput,
			Name: `prefix:"..."`,
			Does: "Is text in front of each value of the field, such as @fact. in front of a fact key.",
			When: "When a person types the value with the prefix.",
		},
		{
			Kind: KindOutput,
			Name: `table:"-"`,
			Does: "Leaves the field out of a table. The key-value block of a get shows it.",
			When: "On a field that a list does not need and a get does, such as a reason or a time of each change.",
		},
		{
			Kind: KindOutput,
			Name: `title:"..."`,
			Does: "On the field _ struct{} of the output type: a different heading. The blank field is not in the JSON or the spec.",
			When: "When the name of the group is not what the output is, such as Stale deployments in the file group.",
		},

		{
			Kind: KindDeps,
			Name: "avfs.VFS",
			Does: "The filesystem, for a command that reads or writes files on the machine where it runs, such as a key or the config. cmd/commands.go gives the real one. A test gives memfs.New(), an in-memory filesystem, and reads what the command wrote from it.",
			When: "For each file that a command reads or writes. Never the os package, which a test cannot replace.",
		},
		{
			Kind: KindDeps,
			Name: "fs.FS",
			Does: "A read-only tree of files, such as the skills and templates that the binary embeds. A test gives fstest.MapFS with the files that it needs.",
			When: "For files that the binary carries, not files on the host.",
		},
		{
			Kind: KindDeps,
			Name: "config.Config, string",
			Does: "The config that osapi read, and plain values such as the path of the config file, a directory or the version. A test gives the value.",
			When: "For what the command reads and does not change. A value is not a service, so it needs no interface.",
		},
		{
			Kind: KindDeps,
			Name: "*slog.Logger",
			Does: "The logger of the process. A test gives slog.New(slog.DiscardHandler).",
			When: "For a process, such as osapi agent start, that logs while it runs. A command that prints its result returns it as its output instead.",
		},
		{
			Kind: KindDeps,
			Name: "func(...)",
			Does: "A function that the CLI gives, such as Behaviors, Validate or NewAgent: it reads the registry of this binary, or makes a part of the process. A test gives a function that returns what the case needs.",
			When: "For one call, or a constructor. When a test must check the calls that a part gets, make the part an interface, as Agent is, and return it from the function.",
		},
		{
			Kind: KindDeps,
			Name: "an interface of the group, in deps.go",
			Does: "A service with methods, such as Tokens, or a part that a constructor returns, such as Agent, Controller and Server. The tests give the mockgen mock from the group's mocks/, from go generate ./internal/cli/commands/<group>/mocks.",
			When: "For a service whose calls a test asserts. Never write a fake by hand: a fake does not fail on a call that the test did not expect.",
		},
		{
			Kind: KindDeps,
			Name: "*cobra.Command",
			Does: "The root command of the CLI, for a group that reads the tree of commands, such as cli docs and cli command.",
			When: "Only to read the commands of this binary.",
		},
		{
			Kind: KindDeps,
			Name: "another group's Deps",
			Does: "The Deps of a different group, for a command that runs that group's work, as osapi start runs the NATS server, the controller and the agent.",
			When: "For a command that is several other commands at one time. Each part keeps its own Deps.",
		},
		{
			Kind: KindWiring,
			Name: "Provide[Deps](Group, fn)",
			Does: "Gives the group's methods their Deps: fn makes them when a command runs, after osapi read its config. A command of a group whose Deps has a field, and that has no Provide, fails.",
			When: "In cmd/commands.go, one time for each group whose Deps has a field, with the real config, file system or client.",
		},

		{
			Kind: KindCheck,
			Name: "Flaggable(type)",
			Does: "Whether a flag holds a value of the type: string, bool, int, []string, map[string]string or []byte, a file. The command generator builds a flag of each, and behavior validate asks it of each CLI behavior's input, so the two agree.",
			When: "Give an input field one of those types.",
		},
		{
			Kind: KindCheck,
			Name: "TagProblems(type)",
			Does: "Each output tag, such as render, that the output does not know, in the type and the types it holds: the check that cli command validate makes of a registered command. behavior validate asks it of each behavior's output, so its author sees the problem in the validator they run.",
			When: "Use a render value that cli command spec get lists.",
		},
		{
			Kind: KindCheck,
			Name: "osapi cli command validate",
			Does: "Checks every local group against the rules of a definition, and the output of every group: each output tag is one the output knows, and each method with no output says what occurred. Each problem names the rule and ends with its fix. validate --help lists each rule. go generate fails on a local group that breaks one.",
			When: "After each change to a group.",
		},
		{
			Kind: KindCheck,
			Name: "TestRefusals",
			Does: "Runs each input that a method's Fails declares through Run, as the generated command does, and expects a usage error.",
			When: "It is in internal/cli/commands/all, and covers each group with no change.",
		},
		{
			Kind: KindCheck,
			Name: "TestCommandShape",
			Does: "Holds each command under osapi client, and each command of a local group, to <resource> <verb>, with the verbs and actions of command.Verbs and command.Actions.",
			When: "It is in cmd/help_style_test.go. A new action goes in command.Actions; a noun does not.",
		},
		{
			Kind: KindCheck,
			Name: "go generate ./internal/cli/commands/all",
			Does: "Writes, for each local group, the cobra commands in gen/cli.go, and internal/cli/commands/all/gen, which registers every group's commands. cmd attaches them under the root.",
			When: "After each change to a group. go-generate-check fails when a gen/ file is stale. Never edit a gen/ file.",
		},
		{
			Kind: KindCheck,
			Name: "osapi cli command create --name <group> --description \"...\"",
			Does: "Writes a new group under internal/cli/commands/<group>: a list and a get, an empty Deps, its types and the tests of its methods, and adds its import to internal/cli/commands/all. It builds, validates and passes its tests; command-scaffold-check in just test holds it to that.",
			When: "To start a new group. Then run go generate ./internal/cli/commands/all and just cli-docs-generate, and change it.",
		},
		{
			Kind: KindCheck,
			Name: "osapi cli command get --name <group>",
			Does: "Describes a group, local, endpoint or behavior: each command and its one line of help. A local group's commands also have their flags and the inputs they refuse.",
			When: "To see a group's commands, and the flags and refusals of a local group.",
		},
		{
			Kind: KindCheck,
			Name: "osapi cli command list",
			Does: "Lists each command of the binary, and the layer that made it: command, endpoint, behavior, or hand-written.",
			When: "To see what a definition makes, and what is still written by hand in cmd/.",
		},

		{
			Kind: KindGenerated,
			Name: "internal/cli/commands/<group>/gen/",
			Does: "The cobra commands of the group, from go generate ./internal/cli/commands/all.",
			When: "Change the definition, then generate. A generated file that is wrong is fixed in the definition or in internal/cli/command/gen.",
		},
		{
			Kind: KindGenerated,
			Name: "internal/cli/commands/<group>/mocks/",
			Does: "The doubles of the group's Deps, from go generate ./internal/cli/commands/<group>/mocks.",
			When: "Run it again after deps.go changes.",
		},
		{
			Kind: KindGenerated,
			Name: "internal/cli/commands/all/gen/",
			Does: "The import of each group's gen package, so the CLI registers its commands, from go generate ./internal/cli/commands/all.",
			When: "Change the definition, then generate.",
		},
		{
			Kind: KindGenerated,
			Name: "docs/docs/sidebar/usage/cli/",
			Does: "The CLI pages, from just cli-docs-generate. The page of a local command comes from .Short, .Long, .Example, .Prints and each input field's doc.",
			When: "Change the definition, then generate. Do not write a CLI page by hand.",
		},
	},
}
