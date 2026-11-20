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

package startcmd

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/agentcmd"
	"github.com/osapi-io/osapi/internal/cli/commands/controllercmd"
	"github.com/osapi-io/osapi/internal/cli/commands/natsserver"
	"github.com/osapi-io/osapi/internal/job"
)

// Start checks the host and starts one tracer. Then it makes the NATS
// server, which the others connect to, the controller and the agent, with
// the namespace of the controller for the job subjects. The agent's
// connection closes, then the tracer stops, after the components stop.
func Start(
	ctx context.Context,
	d Deps,
) (command.Running, error) {
	if err := d.Agent.CheckHost(); err != nil {
		return command.Running{}, fmt.Errorf("unsupported host: %w", err)
	}
	stop, err := d.Agent.Tracer(ctx, "osapi")
	if err != nil {
		return command.Running{}, fmt.Errorf("start the tracer: %w", err)
	}
	job.Init(d.Controller.Config.Controller.NATS.Namespace)

	n, err := natsserver.Components(ctx, d.NATS)
	if err != nil {
		return command.Running{}, err
	}
	c, err := controllercmd.Components(ctx, d.Controller)
	if err != nil {
		return command.Running{}, err
	}
	a, err := agentcmd.Components(ctx, d.Agent)
	if err != nil {
		return command.Running{}, err
	}

	all := command.Running{
		Components: append(append(c.Components, a.Components...), n.Components...),
		Cleanup:    a.Cleanup,
	}

	return command.Then(all, func(ctx context.Context) { _ = stop(ctx) }), nil
}
