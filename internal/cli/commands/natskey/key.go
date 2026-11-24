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

// Package natskey is the key group under osapi nats: osapi nats key create,
// which makes the key that encrypts the JetStream store of the embedded
// server. It is a local command: it writes a file, and calls no API.
package natskey

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the key group's definition.
var Group = command.Register(command.Define(
	"key", "Make the key that encrypts the JetStream store on this host",
	command.Under(command.ParentNATS),
	command.Uses[Deps](),
	command.Method[CreateInput, CreateOutput]("create").
		Short("Make the key that encrypts the JetStream store").
		Long("The command writes 32 random bytes as hex to the file that\n"+
			"`nats.server.jetstream.key_file` names, with mode 0600. It makes the\n"+
			"directory of the file with mode 0700. The embedded NATS server reads the\n"+
			"key when it starts, and encrypts each stream and bucket in `store_dir`.\n\n"+
			"The command refuses to write a file that exists, because a new key in\n"+
			"place of the old key makes each stream unreadable. Keep a copy of the key\n"+
			"apart from the backup of `store_dir`: without the key, the server cannot\n"+
			"read the store. Refer to [Configuration](../../../configuration.md).").
		Example("-f /etc/osapi/osapi.yaml").
		Prints(sampleCreate).
		Handle(Create),
))
