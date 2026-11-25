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

// Package natsperms is the permissions group under osapi nats: osapi nats
// permissions get, which prints the NATS subjects that one osapi process
// needs, for a NATS server that osapi does not run (osapi-io/osapi#1415).
package natsperms

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the permissions group's definition.
var Group = command.Register(command.Define(
	"permissions", "Print the NATS subjects that each osapi process needs",
	command.Under(command.ParentNATS),
	command.Uses[Deps](),
	command.Method[GetInput, GetOutput]("get").
		Short("Print the NATS subjects that one osapi process needs").
		Long("It prints the least that the role can publish and subscribe to, with the\n"+
			"streams and buckets of the config and the namespace of `controller.nats`.\n"+
			"Give these to the NATS user of the role on a NATS server that osapi does\n"+
			"not run. `nsc` is the flags of `nsc add user` or `nsc edit user`, for a\n"+
			"server in operator mode. The embedded server applies them itself to a user\n"+
			"with a `role` in `nats.server.auth.users`.").
		Example("--role agent").
		Prints(sampleGet).
		Example("--role controller --json").
		Fails("no role", GetInput{}).
		Fails("a role that osapi does not have", GetInput{Role: "admin"}).
		Handle(Get),
))
