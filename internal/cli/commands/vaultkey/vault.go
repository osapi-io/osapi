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

// Package vaultkey is the vault group under osapi controller: osapi
// controller vault key create and rotate, which make and rotate the key of a
// nats vault. They are local commands: the key is a file on the controller
// host, and the API does not give it or change it.
package vaultkey

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the vault group's definition.
var Group = command.Register(command.Define(
	"vault", "Work on the vaults of the controller on this host",
	command.Under(command.ParentController),
	command.Uses[Deps](),
	command.SubDoc("key", "Make and rotate the key of a nats vault"),
	command.Method[Input, CreateOutput]("create").
		Sub("key").
		Short("Make the key of a nats vault").
		Long("The command writes a new random AES-256 key to the `key_file` of the\n"+
			"vault in `controller.secrets.vaults`, with the mode 0600. The vault\n"+
			"encrypts each value with this key. The command does not replace a key that\n"+
			"exists, because each value under it would then not decrypt. Use `rotate`\n"+
			"for that. It prints the ID of the key, not the key.").
		Example("--name default").
		Prints(sampleCreate).
		Fails("no name", Input{}).
		Handle(Create),
	command.Method[Input, RotateOutput]("rotate").
		Sub("key").
		Short("Rotate the key of a nats vault").
		Long("The command adds a new key to the `key_file` of the vault, and keeps the\n"+
			"old key for reads. Then it connects to NATS as the controller does,\n"+
			"encrypts each value of the vault again with the new key, and removes the\n"+
			"old key from the file. A controller that runs reads the file at each call,\n"+
			"so it needs no restart. If the rotation stops, the file keeps the old key,\n"+
			"and a second rotation completes it. The earlier versions of a value stay\n"+
			"under the old key, and do not decrypt after the rotation.").
		Example("--name default").
		Prints(sampleRotate).
		Fails("no name", Input{}).
		Handle(Rotate),
))
