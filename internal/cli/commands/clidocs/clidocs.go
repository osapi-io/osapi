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

// Package clidocs is the docs group under osapi cli: osapi cli docs
// generate, which writes the CLI reference pages of the docs site.
package clidocs

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the docs group's definition.
var Group = command.Register(command.Define(
	"docs", "Write the CLI reference pages of the docs site",
	command.Under(command.ParentCLI),
	command.Uses[Deps](),
	command.SubDoc("dsl", "Write the DSL reference pages of the docs site"),
	command.SubDoc("reference", "Write the reference pages of the docs site"),
	command.Method[GenerateInput, GenerateOutput]("generate").
		Short("Write the CLI reference pages of the docs site").
		Long("It writes a page for each command of this binary into `--dir`, and\n"+
			"removes each page there that no command has. `just cli-docs-generate`\n"+
			"runs it, and `just test` checks that the pages match the binary.").
		Example("").
		Prints(sampleGenerate).
		Example("--dir /tmp/cli").
		Handle(Generate),
	command.Method[DSLInput, DSLOutput]("generate").
		Sub("dsl").
		Short("Write the DSL reference pages of the docs site").
		Long("It writes a page for each DSL of this binary into `--dir`: the behavior\n"+
			"DSL, chains, the endpoint DSL and the command DSL. Each page lists what the\n"+
			"`spec get` command of the DSL prints, from the same spec. It removes each\n"+
			"generated page there that no DSL has, and keeps a page that a person wrote.\n"+
			"`just cli-docs-generate` runs it, and `just test` checks that the pages\n"+
			"match the binary.").
		Example("").
		Prints(sampleDSL).
		Example("--dir /tmp/dsl").
		Handle(GenerateDSL),
	command.Method[ReferenceInput, ReferenceOutput]("generate").
		Sub("reference").
		Short("Write the reference pages of the docs site").
		Long("It writes a page for each behavior, chain, endpoint group and local group\n"+
			"of this binary under `docs/docs/sidebar/reference`. Each page is the output\n"+
			"of the `list` or `get` command that shows it, such as `osapi behavior get\n"+
			"--name dns`, printed as markdown. Thus a page shows what its command shows.\n"+
			"It removes each generated page there that no command writes now. `just\n"+
			"cli-docs-generate` runs it, and `just test` checks that the pages match the\n"+
			"binary.").
		Example("").
		Prints(sampleReference).
		Example("--root ~/src/osapi").
		Handle(GenerateReference),
))
