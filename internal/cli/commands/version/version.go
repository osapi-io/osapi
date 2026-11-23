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
// Package version is the version group: osapi version, which prints the
// build of the binary.
package version

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the version group's definition. It is flat, so its method is a
// command of the root: osapi version.
var Group = command.Register(command.Define(
	"version", "Display the version of osapi",
	command.Flat(),
	command.Uses[Deps](),
	command.Method[Input, Info]("version").
		Short("Display the version of osapi").
		Long("The build of this binary: the version that each component reports, the\n"+
			"commit and the tree it was built from, and the Go that built it.").
		Example("").
		Prints(sampleInfo).
		Example("--json").
		Handle(Version),
))
