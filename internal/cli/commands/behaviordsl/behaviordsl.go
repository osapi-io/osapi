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

// Package behaviordsl is the behavior group: the commands that list,
// describe, scaffold, teach and check the behaviors of this binary. The
// package is not named behavior, which is the DSL itself.
package behaviordsl

import (
	"github.com/osapi-io/osapi/internal/behavior"
	// Each behavior registers itself, so these commands read them.
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/guide"
)

// Group is the behavior group's definition.
var Group = command.Register(command.Define(
	"behavior",
	"List, describe and validate the behaviors this binary has",
	command.Long("A behavior is one thing that osapi manages on a host, for example kernel\n"+
		"parameters. [Behaviors](../../../architecture/behaviors.md) tells more.\n"+
		"These commands read the definitions in the binary. Thus they need no config\n"+
		"file and no controller. What they print always agrees with the build that\n"+
		"you run."),
	command.Uses[Deps](),
	command.SubDoc("method", "Show one method of a behavior in full"),
	command.SubDoc(
		"spec",
		"Print the behavior DSL: each option, method, example part and field tag",
	),
	command.Method[ListInput, ListOutput]("list").
		Short("List the behaviors this binary has").
		Long("An internal behavior has no API or CLI of its own. Only other behaviors\n"+
			"use it.").
		Example("").
		Example("--json").
		Handle(List),
	command.Method[GetInput, GetOutput]("get").
		Short("Show a behavior: its surfaces, chain, methods, types and telemetry").
		Long("It shows where osapi serves the behavior, its platforms, the programs that\n"+
			"it runs with sudo, the behaviors that it uses and that use it, its methods\n"+
			"with their routes and permissions, its types, its settings and its\n"+
			"telemetry.\n\n"+
			"**Uses** lists the behaviors this one is built on. The agent builds their\n"+
			"providers first, and the methods of this behavior call theirs on the same\n"+
			"host, in the same job. If a behavior that it uses cannot run on a\n"+
			"platform, this one reports `skipped` there too. **Used by** lists the\n"+
			"behaviors built on this one. A behavior with `Internal: yes` has no route,\n"+
			"no command and no permission. Only the behaviors under **Used by** call\n"+
			"it. `osapi behavior method get` shows one method in full: its fields with\n"+
			"their constraints, its output and its examples.").
		Example("--name sysctl").
		Example("--name sysctl --json").
		Fails("no name", GetInput{}).
		Handle(Get),
	command.Method[MethodInput, MethodOutput]("get").
		Sub("method").
		Short("Show one method of a behavior: its fields, constraints, output and examples").
		Long("Each input field is a flag of its command and a field of its request.\n"+
			"`--json` has the JSON Schema of the input and the output, with each\n"+
			"constraint that osapi checks before it queues the job.").
		Example("--name sysctl --method update").
		Example("--name sysctl --method update --json").
		Fails("no name", MethodInput{Method: "get"}).
		Fails("no method", MethodInput{Name: "sysctl"}).
		Handle(MethodGet),
	command.Method[CreateInput, CreateOutput]("create").
		Short("Write a new behavior's standard layout in an osapi checkout").
		Long("Before you change anything, the result builds, validates and passes every\n"+
			"test. Then change the types, the provider and the examples to what the\n"+
			"behavior manages. Writes nothing when the name is taken. A behavior that\n"+
			"belongs to a group, such as `network`, is nested under it: its routes are\n"+
			"`/api/node/{hostname}/network/<name>`, its commands `osapi client node\n"+
			"network <name>`, and its directory `internal/behaviors/network/<name>/`.\n"+
			"[Your first behavior](../../../development/your-first-behavior.md) walks\n"+
			"through the rest.\n\n"+
			"A new behavior is one an operator calls, with an API and commands. Work\n"+
			"that two behaviors share, such as writing a file, becomes an internal\n"+
			"behavior that the two require: add `behavior.Internal` to it then, and\n"+
			"`behavior.Requires` to each user. `osapi behavior validate` fails an\n"+
			"internal behavior that no behavior requires.").
		Example(`--name motd --description "The message shown at login"`).
		Example(`--name motd --description "The message shown at login" --root ~/src/osapi`).
		Example(`--group network --name neighbors --description "The host's ARP neighbors"`).
		Fails("no name", CreateInput{Description: "x"}).
		Fails("no description", CreateInput{Name: "x"}).
		Handle(Create),
	command.Method[SpecInput, SpecOutput]("get").
		Sub("spec").
		Short("Print the behavior DSL: each option, method, example part and field tag").
		Long("Print each part of the behavior DSL, what it makes osapi do and when to use\n"+
			"it. The list comes from this binary. Thus it agrees with the build that you\n"+
			"run. Read it before you write a behavior, then check the result with\n"+
			"`osapi behavior validate`.").
		Example("").
		Example("--json").
		Handle(Spec),
	command.Method[ValidateInput, ValidateOutput]("validate").
		Short("Check every behavior against the rules the platform relies on").
		Long("It prints each problem with its behavior, method, rule and message, then\n"+
			"the fix of each rule, and exits 1 when a rule fails. `osapi behavior spec\n"+
			"get` describes each part of a behavior.\n\n"+
			"The rules:\n\n"+behavior.Checks.Help()+"\n"+
			"`go test` runs the examples and refusals of each behavior on a fake host,\n"+
			"and checks these. Each failure ends with its fix:\n\n"+behavior.TestChecks.Help()+
			"\n`go test` also checks the chain rules that need a run of the examples.\n"+
			"`osapi behavior chain validate` checks them too:\n\n"+runChainChecks().Help()).
		Example("").
		Example("--json").
		Handle(Validate),
))

// runChainChecks are the chain rules that need a run of the examples, which
// go test checks for each behavior.
func runChainChecks() guide.Checks {
	var out guide.Checks
	for _, c := range behavior.ChainChecks {
		if c.Rule == behavior.RuleCalled || c.Rule == behavior.RuleUsed {
			out = append(out, c)
		}
	}

	return out
}
