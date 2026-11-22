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

// Package agentcmd is the agent group: osapi agent start, which runs the
// agent on a managed host. The package is not named agent, which is the
// agent itself.
package agentcmd

import "github.com/osapi-io/osapi/internal/cli/command"

// Group is the agent group's definition. It is flat, Under agent, so its
// method is osapi agent start.
var Group = command.Register(command.Define(
	"agent", "Run the agent",
	command.Flat(),
	command.Under(command.ParentAgent),
	command.Uses[Deps](),
	command.Process("start").
		Short("Start the node agent").
		Long("The agent connects to NATS and subscribes to the job streams. It processes\n"+
			"each job when the job arrives. It uses a provider for its platform to run\n"+
			"each operation, for example node status, DNS queries and ping.").
		Example("").
		Example("--agent-max-jobs 20").
		Flag(
			"agent.nats.host",
			"agent.nats.port",
			"agent.nats.client_name",
			"agent.queue_group",
			"agent.hostname",
			"agent.max_jobs",
			"agent.consumer.max_deliver",
			"agent.consumer.ack_wait",
			"agent.consumer.max_ack_pending",
			"agent.consumer.replay_policy",
			"agent.consumer.back_off",
		).
		Setup(Start),
))
