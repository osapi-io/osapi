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

// Package natssetup is the setup group under osapi nats: osapi nats setup,
// which creates osapi's streams and buckets on a NATS server that osapi
// does not run, such as an existing cluster.
package natssetup

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the setup group's definition. It is flat, Under nats, so its
// method is osapi nats setup.
var Group = command.Register(command.Define(
	"setup", "Create osapi's streams and buckets on a NATS server",
	command.Flat(),
	command.Under(command.ParentNATS),
	command.Uses[Deps](),
	command.Method[Input, Output]("setup").
		Short("Create osapi's streams and buckets on a NATS server").
		Long("It connects as the controller does, with `controller.nats`: its host,\n"+
			"port, namespace and auth. Run it one time before you start the controller\n"+
			"and the agents with that server.\n\n"+
			"Run it again after you change the settings of a stream or a bucket. Each\n"+
			"stream and bucket is created or updated to match the config, so running it\n"+
			"again changes nothing. `osapi nats start` does the same for the\n"+
			"embedded server when the server starts. Thus a deployment that runs the\n"+
			"NATS server of osapi does not need this command.").
		Example("-f /etc/osapi/osapi.yaml").
		Prints(sampleSetup).
		Handle(Setup),
))
