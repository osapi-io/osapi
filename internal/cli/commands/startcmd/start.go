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

// Package startcmd is the start group: osapi start, which runs the NATS
// server, the controller and the agent in one process.
package startcmd

import (
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/config"
)

// Group is the start group's definition. It is flat at the root, so its
// method is osapi start.
var Group = command.Register(command.Define(
	"start", "Start all components (NATS, controller, agent)",
	command.Flat(),
	command.Uses[Deps](),
	command.Config("telemetry", config.Telemetry{}),
	command.Process("start").
		Short("Start all components (NATS, controller, agent)").
		Long("This runs the embedded NATS server, the API server and the agent together.\n"+
			"Use it for a deployment on one host. For a deployment on more than one\n"+
			"host, start the NATS server, the controller and each agent alone.\n\n"+
			"The NATS server starts first, and waits until it is ready. Then osapi\n"+
			"makes the streams and buckets, starts the API server and starts the agent.\n"+
			"On `SIGINT` or `SIGTERM`, the three components stop at the same time, in\n"+
			"10 seconds or less. Then osapi closes the NATS connections and the\n"+
			"telemetry exporters.").
		Example("").
		Logs(startLogs...).
		Setup(Start),
))
