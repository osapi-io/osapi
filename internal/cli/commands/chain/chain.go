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

// Package chain is the chain group of commands, osapi dsl chain: the
// behaviors that a behavior calls on the same host, in the same job.
package chain

import (
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli/command"
)

// Group is the chain group's definition: osapi dsl chain. A chain is
// the behaviors that a behavior calls on the same host, in the same job,
// and the ones that those call.
var Group = command.Register(command.Define(
	"chain",
	"Show and check the chains of behaviors that a behavior calls",
	command.Under(command.ParentDSL),
	command.Long("A behavior that requires others calls them on the same host, in the same\n"+
		"job: dns calls netplan, and netplan calls managedfile. That is a chain, named\n"+
		"by the behavior at its top. `behavior.Requires` on each behavior makes it;\n"+
		"it is not declared on its own. These commands read the definitions in the\n"+
		"binary, and run the examples on the fake host that `go test` uses. Thus they\n"+
		"need no config file and no controller, and show what the code does now."),
	command.Uses[Deps](),
	command.SubDoc("spec", "Print how to build a chain, and the building blocks of this binary"),
	command.Method[ListInput, ListOutput]("list").
		Short("List each behavior in a chain: what it calls, and what calls it").
		Long("**Chain** is the tree of the behaviors that a behavior calls, down to the\n"+
			"last one. A building block, with `Internal`, is in the list for the\n"+
			"behaviors that call it.").
		Example("").
		Prints(sampleList).
		Example("--json").
		Handle(List),
	command.Method[GetInput, GetOutput]("get").
		Short("Show a chain: its tree, and what each call of each example sends and gets").
		Long("It runs each example of the behavior on the fake host that `go test` uses,\n"+
			"and records each `behavior.Call`: the behavior, the method, the input, and\n"+
			"the output or the error. The table cuts each input and output; `--json`\n"+
			"has them in full, under `runs`. An input field tagged `secret` shows as\n"+
			"`[redacted]`.").
		Example("--name dns").
		Prints(sampleGet).
		Example("--name dns --json").
		Fails("no name", GetInput{}).
		Handle(Get),
	command.Method[ValidateInput, ValidateOutput]("validate").
		Short("Check every chain against the chain rules").
		Long("It prints each problem with its behavior, method, rule and message, then\n"+
			"the fix of each rule, and exits 1 when a rule fails. `osapi dsl behavior\n"+
			"validate` checks the rules that hold when the behaviors are built, and\n"+
			"`go test` checks the ones that need a run of the examples.\n\n"+
			"The rules:\n\n"+behavior.ChainChecks.Help()).
		Example("").
		Prints(sampleValidate).
		Example("--json").
		Handle(Validate),
	command.Method[SpecInput, SpecOutput]("get").
		Sub("spec").
		Short("Print how to build a chain, and the building blocks of this binary").
		Long("Print what a chain is, how a behavior builds on others and calls them, what\n"+
			"happens when a step fails, and each building block of this binary. Read it\n"+
			"before a behavior requires another, then check the result with `osapi\n"+
			"dsl chain validate`.").
		Example("").
		Prints(sampleSpec).
		Example("--json").
		Handle(Spec),
))
