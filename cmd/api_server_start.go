// Copyright (c) 2024 John Dewey

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
	"time"

	"github.com/labstack/echo/v4"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/api"
	jobclient "github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/messaging"
)

// ServerManager responsible for Server operations.
type ServerManager interface {
	Lifecycle
	// CreateHandlers initializes handlers and returns a slice of functions to register them.
	CreateHandlers(
		jobClient jobclient.JobClient,
	) []func(e *echo.Echo)
	// RegisterHandlers registers a list of handlers with the Echo instance.
	RegisterHandlers(handlers []func(e *echo.Echo))
}

// apiServerStartCmd represents the apiServerStart command.
var apiServerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long: `Start the API server.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		// Create NATS client for job system
		var nc messaging.NATSClient = natsclient.New(logger, &natsclient.Options{
			Host: appConfig.Job.Client.Host,
			Port: appConfig.Job.Client.Port,
			Auth: natsclient.AuthOptions{
				AuthType: natsclient.NoAuth,
			},
			Name: appConfig.Job.Client.ClientName,
		})

		if err := nc.Connect(); err != nil {
			logFatal("failed to connect to NATS for job client", err)
		}

		jobsKV, err := nc.CreateKVBucket(appConfig.Job.KVBucket)
		if err != nil {
			logFatal("failed to create KV bucket", err)
		}

		jc, err := jobclient.New(logger, nc, &jobclient.Options{
			Timeout:  30 * time.Second,
			KVBucket: jobsKV,
		})
		if err != nil {
			logFatal("failed to create job client", err)
		}

		var sm ServerManager = api.New(appConfig, logger)
		handlers := sm.CreateHandlers(jc)
		sm.RegisterHandlers(handlers)

		sm.Start()
		runServer(ctx, sm, func() {
			if natsConn, ok := nc.(*natsclient.Client); ok && natsConn.NC != nil {
				natsConn.NC.Close()
			}
		})
	},
}

func init() {
	apiServerCmd.AddCommand(apiServerStartCmd)
}
