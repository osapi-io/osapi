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

// Package agentkey is the key group under osapi agent: osapi agent key
// fingerprint, which shows what an operator needs to admit this machine.
package agentkey

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the agent key group's definition.
var Group = command.Register(command.Define(
	"key", "Read the keys of the agent on this host",
	command.Under(command.ParentAgent),
	command.Uses[Deps](),
	command.Method[FingerprintInput, FingerprintOutput]("fingerprint").
		Short("Show the fingerprint of the agent's keys, and what admits this machine").
		Long("Run this on the machine. It reads the two agent keys in `agent.pki.key_dir`\n"+
			"(`/etc/osapi/pki/agent` by default): the Ed25519 signing key and the X25519\n"+
			"encryption key. It prints the machine ID, the hostname, the public\n"+
			"encryption key, the fingerprint that covers the two keys, and the command\n"+
			"that admits the machine. The controller admits the machine only when the\n"+
			"fingerprint is the one of the keys in its request. The command makes each\n"+
			"key that does not exist.").
		Example("").
		Prints(sampleFingerprint).
		Handle(Fingerprint),
))
