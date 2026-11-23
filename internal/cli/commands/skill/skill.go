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

// Package token is the token group of local commands: osapi token generate,
// roles list and validate, which sign and read the bearer tokens of the API
// with the signing key of the config. They run in the CLI, and call no API.
// Package skill is the skill group: osapi skill list and get, which show the
// agent skills that this binary has.
package skill

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the skill group's definition.
var Group = command.Register(command.Define(
	"skill", "Show the agent skills that this binary has",
	command.Uses[Deps](),
	command.Method[ListInput, ListOutput]("list").
		Short("List the skills that this binary has").
		Example("").
		Prints(sampleList).
		Example("--json").
		Handle(List),
	command.Method[GetInput, GetOutput]("get").
		Short("Print a skill, or one of its reference files").
		Long("It prints the file as it is. `--json` has the name, the reference and the\n"+
			"content.").
		Example("--name build-a-behavior").
		Prints(sampleGet).
		Example("--name build-a-behavior --reference references/services.md").
		Fails("no name", GetInput{}).
		Handle(Get),
))
