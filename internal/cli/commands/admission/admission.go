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

// Package admission is the admission group under osapi agent: osapi agent
// admission get, which shows what an operator needs to admit this machine.
package admission

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the admission group's definition.
var Group = command.Register(command.Define(
	"admission", "Show what admits this machine",
	command.Under(command.ParentAgent),
	command.Uses[Deps](),
	command.Method[GetInput, GetOutput]("get").
		Short("Show what admits this machine").
		Long("Run this on the machine. It prints the machine ID, the hostname and the\n"+
			"fingerprint of the agent key, in `agent.pki.key_dir` (`/etc/osapi/pki` by\n"+
			"default), and the command that admits the machine. The controller admits\n"+
			"the machine only when the fingerprint is the one of the key in its\n"+
			"request. If no key exists, the command makes a new key pair.").
		Example("").
		Prints(sampleGet).
		Handle(Get),
))
