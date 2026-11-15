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
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/job"
)

// natsSetupCmd creates osapi's streams and buckets on a NATS server osapi does
// not run, such as an existing cluster, through the controller's connection.
// `nats server start` does the same for its embedded server.
var natsSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create osapi's streams and buckets on a NATS server",
	Long: `Create the JetStream streams and key-value buckets osapi needs on a NATS
server it does not run, such as an existing cluster. It connects as the
controller does, with controller.nats: its host, port, namespace and auth, so
a server in operator (JWT) mode is reached with the controller's credentials.

Each stream and bucket is created or updated to match the config, so running
it again changes nothing. osapi nats server start does the same for its
embedded server, so this is needed only when the server is someone else's.`,
	Example: `  osapi nats setup -f /etc/osapi/osapi.yaml`,
	Run: func(cmd *cobra.Command, _ []string) {
		n := appConfig.Controller.NATS
		job.Init(n.Namespace)
		log := logger.With("component", "nats")

		if err := bus.Prepare(cmd.Context(), log, &natsclient.Options{
			Host: n.Host,
			Port: n.Port,
			Auth: bus.AuthOptions(n.Auth),
			Name: "osapi-nats-setup",
		}, n.Namespace, appConfig.NATS); err != nil {
			cli.LogFatal(log, "failed to set up JetStream infrastructure", err)
		}

		log.Info(
			"JetStream infrastructure ready",
			"host", n.Host,
			"port", n.Port,
			"namespace", n.Namespace,
			"auth.type", n.Auth.Type,
		)
	},
}

func init() {
	natsCmd.AddCommand(natsSetupCmd)
}
