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

// Package clidocs is the docs group under osapi dsl: osapi dsl docs
// generate, which writes the generated pages of the docs site.
package clidocs

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the docs group's definition.
var Group = command.Register(command.Define(
	"docs", "Write the generated pages of the docs site",
	command.Under(command.ParentDSL),
	command.Uses[Deps](),
	command.Method[GenerateInput, GenerateOutput]("generate").
		Short("Write the CLI, DSL and reference pages of the docs site").
		Long("It writes each generated page of the docs site from this binary, and\n"+
			"removes each generated page that no command, DSL or group has now:\n\n"+
			"- a page for each command, under `docs/docs/sidebar/usage/cli`;\n"+
			"- a page for each DSL, under `docs/docs/sidebar/development/dsl`, from the\n"+
			"  spec that the `spec get` command of the DSL prints;\n"+
			"- a page for each behavior, chain, endpoint group and local group, under\n"+
			"  `docs/docs/sidebar/reference`, from the output of its `list` or `get`\n"+
			"  command, such as `osapi dsl behavior get --name dns`.\n\n"+
			"A DSL page that a person wrote stays. `just cli-docs-generate` runs it, and\n"+
			"`just test` checks that the pages match the binary.").
		Example("").
		Prints(sampleGenerate).
		Example("--root ~/src/osapi").
		Handle(Generate),
))
