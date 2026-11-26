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

// Package join is the join group under osapi controller: osapi controller
// join token create, list and revoke, which make, show and remove the tokens
// that admit an agent at its join (osapi-io/osapi#1313). They are local
// commands: they connect to NATS as the controller does, and the API does
// not give a token.
package join

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the join group's definition.
var Group = command.Register(command.Define(
	"join", "Make and remove the tokens that admit an agent at its join",
	command.Under(command.ParentController),
	command.Uses[Deps](),
	command.SubDoc("token", "Make, list and revoke the join tokens"),
	command.Method[CreateInput, CreateOutput]("create").
		Sub("token").
		Short("Make a join token, and print it with the join string").
		Long("The token admits an agent at its join, until it expires or its uses are\n"+
			"gone. The controller keeps a hash of it, not the token: this command prints\n"+
			"the token one time.\n\n"+
			"By default the agent joins over NATS, and needs to reach NATS only. The join\n"+
			"string has the address of NATS (`controller.join.url`, or else\n"+
			"`nats.server.url`), the bootstrap user, which can only send the join request,\n"+
			"and for a `tls://` server the fingerprint of its certificate, which the agent\n"+
			"pins. The controller gives the agent its own NATS user at the join.\n\n"+
			"With `--transport https`, the agent joins over the controller's API: the join\n"+
			"string has `controller.api.url` and the fingerprint of the API's certificate.\n"+
			"It needs `controller.api.tls.enabled`.").
		Example("").
		Prints(sampleCreate).
		Check("controller join token list", "", "items.[id={out.id}].uses_left", 1).
		Example("--expires 24h --uses 10 --labels group=web").
		Check("controller join token list", "",
			"items.[id={out.id}].uses_left", 10, "items.[id={out.id}].labels.group", "web").
		Example("--transport https").
		Check("controller join token list", "", "items.[id={out.id}].uses_left", 1).
		Fails("no expiry", CreateInput{Uses: 1, Transport: "nats"}).
		Fails("a duration that is not one", CreateInput{Expires: "soon", Uses: 1, Transport: "nats"}).
		Fails("no use", CreateInput{Expires: "1h", Uses: 0, Transport: "nats"}).
		Fails("no transport", CreateInput{Expires: "1h", Uses: 1}).
		Fails("a transport that osapi does not have", CreateInput{Expires: "1h", Uses: 1, Transport: "ssh"}).
		Handle(Create),
	command.Method[ListInput, ListOutput]("list").
		Sub("token").
		Short("List the join tokens, with their expiry and uses left").
		Example("").
		Prints(sampleList).
		Example("--json").
		Handle(List),
	command.Method[RevokeInput, RevokeOutput]("revoke").
		Sub("token").
		Short("Revoke a join token, so it admits no agent").
		Example("--id 1a2b3c4d").
		Prints(sampleRevoke).
		Check("controller join token list", "", "items.[id=1a2b3c4d]", command.Absent).
		Fails("no id", RevokeInput{}).
		Handle(Revoke),
))
