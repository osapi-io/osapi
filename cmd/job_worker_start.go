// Copyright (c) 2025 John Dewey

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

	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/job/worker"
	"github.com/retr0h/osapi/internal/messaging"
)

// jobWorkerStartCmd represents the jobWorkerStart command.
var jobWorkerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long: `Start the job worker.
It processes jobs as they become available.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		// Create NATS client using the nats-client package
		var nc messaging.NATSClient = natsclient.New(logger, &natsclient.Options{
			Host: appConfig.Job.Worker.Host,
			Port: appConfig.Job.Worker.Port,
			Auth: natsclient.AuthOptions{
				AuthType: natsclient.NoAuth,
			},
			Name: appConfig.Job.Worker.ClientName,
		})

		err := nc.Connect()
		if err != nil {
			logFatal("failed to connect to NATS", err)
		}

		// Create/get the jobs KV bucket
		jobsKV, err := nc.CreateKVBucket(appConfig.Job.KVBucket)
		if err != nil {
			logFatal("failed to create KV bucket", err)
		}

		// Create job client
		var jc client.JobClient
		jc, err = client.New(logger, nc, &client.Options{
			Timeout:  30 * time.Second, // Default timeout
			KVBucket: jobsKV,
		})
		if err != nil {
			logFatal("failed to create job client", err)
		}

		// Create provider factory and providers
		providerFactory := worker.NewProviderFactory(logger)
		hostProvider, diskProvider, memProvider, loadProvider, dnsProvider, pingProvider := providerFactory.CreateProviders()

		w := worker.New(
			appFs,
			appConfig,
			logger,
			jc,
			hostProvider,
			diskProvider,
			memProvider,
			loadProvider,
			dnsProvider,
			pingProvider,
		)

		w.Start(ctx)
	},
}

func init() {
	jobWorkerCmd.AddCommand(jobWorkerStartCmd)
}
