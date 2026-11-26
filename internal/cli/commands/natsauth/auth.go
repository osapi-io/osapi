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

// Package natsauth is the auth group under osapi nats: osapi nats auth
// password create and operator create, which make the authentication of a
// NATS setup for osapi, with a credential for each osapi role
// (osapi-io/osapi#1420).
package natsauth

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the auth group's definition.
var Group = command.Register(command.Define(
	"auth",
	"Make the NATS authentication of osapi, a credential for each role",
	command.Under(command.ParentNATS),
	command.Uses[Deps](),
	command.SubDoc("password", "A user and a password for each osapi role"),
	command.SubDoc("nkey", "An nkey for each osapi role"),
	command.SubDoc(
		"operator",
		"An operator, an account and a credentials file for each osapi role",
	),
	command.Method[CreateInput, CreateOutput]("create").
		Sub("password").
		Short("Make a user and a random password for each osapi role").
		Long("It makes a user for each osapi role (setup, controller, agent and join), each\n"+
			"with a random password and the least permissions of its role. It writes\n"+
			"`osapi-nats.yaml`, the part of osapi.yaml with the users of the embedded\n"+
			"server and the auth of the controller and the agent, and `nats-server.conf`,\n"+
			"the authorization block of a NATS server that osapi does not run. Each file\n"+
			"is mode 0600, and no password is on the command line. `--print` also\n"+
			"prints the server's block.").
		Example("--out /etc/osapi/nats").
		Prints(samplePassword).
		Fails("no directory", CreateInput{}).
		Handle(Password),
	command.Method[CreateInput, CreateOutput]("create").
		Sub("operator").
		Short("Make an operator, an account, and a credentials file for each osapi role").
		Long("It makes an operator, its system account, and an account for osapi with\n"+
			"JetStream and a signing key. It writes the server's resolver file\n"+
			"(`resolver.conf`), a credentials file for the setup, controller and agent\n"+
			"roles, each with the least permissions of its role, the signing key that the\n"+
			"controller mints the users of a join with (`signing.nk`), the operator's\n"+
			"seed (`operator.nk`), and `osapi-nats.yaml`, the part of osapi.yaml that uses\n"+
			"them. Each file is mode 0600. Keep `operator.nk` offline: it signs the\n"+
			"accounts. `--print` also prints the resolver file.").
		Example("--out /etc/osapi/nats").
		Prints(sampleOperator).
		Fails("no directory", CreateInput{}).
		Handle(Operator),
	command.Method[CreateInput, CreateOutput]("create").
		Sub("nkey").
		Short("Make an nkey for each osapi role").
		Long("It makes an nkey for the setup, controller and agent roles, and writes the\n"+
			"seed of each (`setup.nk`, `controller.nk`, `agent.nk`), `osapi-nats.yaml`, the\n"+
			"part of osapi.yaml with the nkeys of the embedded server and their roles\n"+
			"(`nkey_roles`), and `nats-server.conf`, the authorization block of a NATS\n"+
			"server that osapi does not run, with the least permissions of each role.\n"+
			"Each file is mode 0600. A join over NATS needs a password or a JWT for its\n"+
			"bootstrap user: use `password create` or `operator create` for it.\n"+
			"`--print` also prints the server's block.").
		Example("--out /etc/osapi/nats").
		Prints(sampleNKey).
		Fails("no directory", CreateInput{}).
		Handle(NKey),
))
