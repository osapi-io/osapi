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

// Package pkikey is the key group under osapi pki: osapi pki key list, get,
// create and rotate, which read, make and rotate the keys of osapi on this
// host.
package pkikey

import (
	"strings"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// Group is the key group's definition.
var Group = command.Register(command.Define(
	"key", "List, make and rotate the keys of osapi on this host",
	command.Under(command.ParentPKI),
	command.Uses[Deps](),
	command.Method[ListInput, ListOutput]("list").
		Short("List the keys, with their state, algorithm and fingerprint or ID").
		Long("Run this on the host. It reads each key of osapi that the config names:\n"+
			"the controller's Ed25519 signing key in controller.pki.key_dir, the agent's\n"+
			"Ed25519 signing key and P-256 encryption key in agent.pki.key_dir, the key\n"+
			"of the JetStream store in nats.server.jetstream.key_file, and the key of each\n"+
			"vault of type nats. A public key shows its fingerprint, and a secret key\n"+
			"shows its ID: never the key. A component that has not started on this host\n"+
			"made no key: its state is not made.").
		Example("").
		Prints(sampleList).
		Handle(List),
	command.Method[GetInput, Key]("get").
		Short("Show one key in full; the agent's key has the command that admits this machine").
		Long("It reads one key, as list does. The agent's signing key also has the\n"+
			"command that admits this machine, with its machine ID and the fingerprint\n"+
			"that covers the agent's two keys. A key that is not on this host is an\n"+
			"error that names osapi pki key create.").
		Example("--name agent").
		Prints(sampleAgent).
		Fails("no name", GetInput{}).
		Fails("a name of more than 64 characters", GetInput{Name: strings.Repeat("a", 65)}).
		Handle(Get),
	command.Method[CreateInput, Key]("create").
		Short("Make a key that is not on this host").
		Long("It makes the key with the name, with mode 0600: the signing key of the\n"+
			"controller, the two keys of the agent, a random key for the JetStream\n"+
			"store, or a random AES-256 key for a vault. It refuses a key that is there,\n"+
			"because a new key in its place is a rotation: a JetStream store and the\n"+
			"values of a vault do not decrypt with it, and an admission does not match\n"+
			"it. The controller and the agent make their keys at their first start, so\n"+
			"run this to make a key before then, such as the JetStream key before NATS\n"+
			"starts. Keep a copy of a JetStream or a vault key apart from the backup of\n"+
			"what it encrypts.").
		Example("--name jetstream").
		Prints(sampleJetStream).
		Fails("no name", CreateInput{}).
		Fails("a name of more than 64 characters", CreateInput{Name: strings.Repeat("a", 65)}).
		Handle(Create),
	command.Method[RotateInput, RotateOutput]("rotate").
		Short("Rotate the key of a vault, and encrypt each value again with it").
		Long("It adds a new key to the key_file of a vault of type nats, encrypts each\n"+
			"value of the vault again with it, then drops the old key. It connects to\n"+
			"NATS as the controller does. A running controller reads the new key at its\n"+
			"next call. A rotation that stops leaves the old key in the file, and a\n"+
			"second rotation completes it. It refuses each other key, with what to do\n"+
			"instead: a new JetStream key makes the store unreadable, and a new signing\n"+
			"key needs each agent, or this machine, admitted again.").
		Example("--name vault:default").
		Prints(sampleRotate).
		Fails("no name", RotateInput{}).
		Fails("a name of more than 64 characters", RotateInput{Name: strings.Repeat("a", 65)}).
		Handle(Rotate),
))
