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
	KindWiring     = "wiring"
	KindCheck      = "check"
	KindGenerated  = "generated"
)

// Spec is every term of the command DSL. osapi cli command spec prints it,
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
		{Kind: KindWiring, Title: "Output and wiring"},
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
			Name: `Under("word", ...)`,
			Does: "Puts the group's command under the command of the words, such as controller for osapi controller endpoint. Attach makes a word that has no command, and uses a hand-written one that cmd/commands.go names.",
			When: "When the commands belong to a command that is there already, such as controller. Without it, the group is under the root.",
		},

		{
			Kind: KindMethod,
			Name: `Method[In, Out]("name")`,
			Does: "A command, osapi <group> <name>, that takes In and prints Out. Each field of In is a flag. Out prints as JSON with --json, and through the method's renderer without it.",
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
			Name: `.Sub("segment")`,
			Does: "Puts the command under a command of its own in the group: osapi <group> <segment> <method>, such as token roles list.",
			When: "When the command is about a noun of the group: the noun is the Sub, and the method is the verb.",
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
			Kind: KindWiring,
			Name: `Render(Group, "method", fn)`,
			Does: "Sets how a method's output prints without --json: fn gets the command, the input and the output. A method without a renderer prints its output as JSON. With --json the output prints as JSON, and the renderer does not run.",
			When: "In render.go of the group, in its init, for each method a person reads at a shell, such as a table or a key-value block.",
		},
		{
			Kind: KindWiring,
			Name: `Acts(Group, "method", fn)`,
			Does: "Sets what a command does with a method's output, with or without --json, such as write it to a file. fn prints its own output.",
			When: "In render.go, for a method whose output is not only printed.",
		},
		{
			Kind: KindWiring,
			Name: "Provide[Deps](Group, fn)",
			Does: "Gives the group's methods their Deps: fn makes them when a command runs, after osapi read its config. A command of a group whose Deps has a field, and that has no Provide, fails.",
			When: "In cmd/commands.go, one time for each group whose Deps has a field, with the real config, file system or client.",
		},

		{
			Kind: KindCheck,
			Name: "osapi cli command validate",
			Does: "Checks every local group against the rules of a definition. Each problem names the rule and ends with its fix. validate --help lists each rule. go generate fails on a group that breaks one.",
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
			Name: "osapi cli command new --name <group> --description \"...\"",
			Does: "Writes a new group under internal/cli/commands/<group>: a list and a get, an empty Deps, its types and the tests of its methods, and adds its import to internal/cli/commands/all. It builds, validates and passes its tests; command-scaffold-check in just test holds it to that.",
			When: "To start a new group. Then run go generate ./internal/cli/commands/all and just cli-docs-generate, and change it.",
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
			Does: "The CLI pages, from just cli-docs-generate. The page of a local command comes from .Short, .Long, .Example and each input field's doc.",
			When: "Change the definition, then generate. Do not write a CLI page by hand.",
		},
	},
}
