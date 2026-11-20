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

// Package commanddsl is the command group under osapi cli: the commands
// that list, describe, scaffold, teach and check the local groups of the
// command DSL. The package is not named for the DSL's own package,
// command, which the generated code imports.
package commanddsl

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the command group's definition.
var Group = command.Register(command.Define(
	"command",
	"List, describe and validate the local commands, and print their DSL",
	command.Under("cli"),
	command.Long("A local command runs in the CLI process and calls no API, such as\n"+
		"`osapi token generate`. Its group is a definition in\n"+
		"`internal/cli/commands/<group>`, and its commands, flags, help and pages\n"+
		"come from it. A command that calls the API is an endpoint\n"+
		"(`osapi controller endpoint`) or a behavior (`osapi behavior`). These\n"+
		"commands read the definitions in the binary. Thus they need no config file\n"+
		"and no running controller."),
	command.Uses[Deps](),
	command.SubDoc(
		"spec",
		"Print the command DSL: each option, method, refinement, field tag and check",
	),
	command.Method[ListInput, ListOutput]("list").
		Short("List each command of this binary, and the layer that made it").
		Long("The layer is `command` for a command of a local group, `endpoint` for a\n"+
			"command of an endpoint group, `behavior` for a command of a behavior, and\n"+
			"`hand-written` for a command in `cmd/` that no definition makes yet.").
		Example("").
		Example("--json").
		Handle(List),
	command.Method[GetInput, GetOutput]("get").
		Short("Describe one group: its commands, and the flags and refusals of a local group").
		Long("`--name` is the command path of the group, such as `token` or `agent`: a\n"+
			"local group, or the group of an endpoint or a behavior. A command of a\n"+
			"local group also has its flags, each with the validate tag of its field,\n"+
			"and the inputs that its `.Fails` declares, which `TestRefusals` runs.").
		Example("--name token").
		Example("--name agent --json").
		Fails("no name", GetInput{}).
		Handle(Get),
	command.Method[CreateInput, CreateOutput]("create").
		Short("Write a new local group that builds and passes its tests").
		Long("It writes `internal/cli/commands/<name>/` in the osapi checkout in the\n"+
			"current directory, and adds its import to `internal/cli/commands/all`. The\n"+
			"group has a list and a get, an empty `Deps` and the tests of its methods.\n"+
			"Then run the steps that it prints. While its `Deps` is empty, the CLI\n"+
			"needs no wiring. `osapi cli command spec get` lists each part of the DSL.").
		Example(`--name widget --description "The widgets of this host"`).
		Fails("no name", CreateInput{Description: "x"}).
		Fails("no description", CreateInput{Name: "x"}).
		Handle(Create),
	command.Method[SpecInput, SpecOutput]("get").
		Sub("spec").
		Short("Print the command DSL: each option, method, refinement, field tag and check").
		Long("Print each part of the command DSL, what it makes osapi do and when to use\n"+
			"it. The list comes from this binary. Thus it agrees with the build that\n"+
			"you run. Read it before you write a local group, then check the result\n"+
			"with `osapi cli command validate`.").
		Example("").
		Example("--json").
		Handle(Spec),
	command.Method[ValidateInput, ValidateOutput]("validate").
		Short("Check every local group against the rules of a definition, and the output of every group").
		Long("It prints each problem with its group, method, rule and message, then the\n"+
			"fix of each rule, and exits 1 when a rule fails. `osapi cli command spec\n"+
			"get` describes each part of a group.\n\n"+
			"The rules:\n\n"+command.Checks.Help()).
		Example("").
		Example("--json").
		Handle(Validate),
))
