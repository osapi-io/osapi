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
// Package initcmd is the init group: osapi init, which installs osapi's agent
// skills into a repository. The package is not named init, which Go keeps.
package initcmd

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the init group's definition. It is flat, so its method is a
// command of the root: osapi init.
var Group = command.Register(command.Define(
	"init", "Install osapi's agent skills into a repository",
	command.Flat(),
	command.Uses[Deps](),
	command.Method[Input, Output]("init").
		Short("Install osapi's agent skills into a repository").
		Long("The command writes the skills in this binary for each `--tool`. It also\n"+
			"writes a short section, with links to the skills, in the instructions file\n"+
			"of that tool: `CLAUDE.md` or `AGENTS.md`. It does not change text outside\n"+
			"the osapi section.\n\n"+
			"Run it again after you upgrade osapi. It writes the skills again to agree\n"+
			"with the binary. It removes the files that an older version installed.\n"+
			"Thus the agent does not read a skill about operations that the binary does\n"+
			"not have now.").
		Example("").
		Prints(sampleInit).
		Example("--tool claude --tool codex").
		Example("--path ./ops-repo --json").
		Fails("an unknown tool", Input{Tools: []string{"vim"}}).
		Handle(Init),
))
