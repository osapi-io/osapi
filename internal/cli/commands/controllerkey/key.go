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

// Package controllerkey is the key group under osapi controller: osapi
// controller key fingerprint, which reads the key pair of the controller on
// this host. It is a local command: it reads a file, and calls no API.
package controllerkey

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the key group's definition.
var Group = command.Register(command.Define(
	"key", "Read the key pair of the controller on this host",
	command.Under(command.ParentController),
	command.Uses[Deps](),
	command.Method[FingerprintInput, FingerprintOutput]("fingerprint").
		Short("Show the fingerprint of the controller's key").
		Long("osapi calculates the fingerprint from the Ed25519 public key of the\n"+
			"controller, in `controller.pki.key_dir` (`/etc/osapi/pki` by default).\n"+
			"Agents use this fingerprint to verify that jobs come from a trusted\n"+
			"controller. If no key exists, the command makes a new key pair. Then it\n"+
			"shows the fingerprint.").
		Example("").
		Handle(Fingerprint),
))
