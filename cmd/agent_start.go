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

package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/telemetry"
)

// agentStartCmd represents the agentStart command.
var agentStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the agent",
	Long: `Start the agent process.
It processes jobs as they become available.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		shutdownTracer, err := telemetry.InitTracer(
			ctx,
			"osapi-agent",
			appConfig.Telemetry.Tracing,
		)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize tracer", err)
		}

		job.Init(appConfig.Agent.NATS.Namespace)

		log := logger.With("component", "agent")
		a, b := setupAgent(ctx, log, appConfig.Agent.NATS)

		a.Start()
		cli.RunServer(ctx, a, func() {
			_ = shutdownTracer(context.Background())
			cli.CloseNATSClient(b.nc)
		})
	},
}

func init() {
	agentCmd.AddCommand(agentStartCmd)
}
