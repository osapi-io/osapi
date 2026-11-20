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

// Package natsserver is the server group under osapi nats: osapi nats
// server start, which runs the embedded NATS server as its own process.
package natsserver

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the server group's definition.
var Group = command.Register(command.Define(
	"server", "Run the embedded NATS server",
	command.Under("nats"),
	command.Uses[Deps](),
	command.Process("start").
		Short("Start the embedded NATS server").
		Long("The server makes the streams, consumers and KV buckets that the job system\n"+
			"needs. `osapi start` runs the same server, together with the controller and\n"+
			"an agent, in one process. Use this command to run the bus as its own\n"+
			"process. The `nats.server` section of the config sets the host, the port and\n"+
			"the authentication. Refer to [Configuration](../../../configuration.md).").
		Example("").
		Setup(Start),
))
