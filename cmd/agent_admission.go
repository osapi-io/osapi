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
	"fmt"

	"github.com/avfs/avfs/vfs/osfs"
	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/pki"
)

// agentAdmissionCmd prints what an operator needs to admit this machine.
var agentAdmissionCmd = &cobra.Command{
	Use:   "admission",
	Short: "Show what admits this machine",
	Long: "Run this on the machine. It prints the machine ID, the hostname and the\n" +
		"fingerprint of the agent key, from the PKI key directory (`/etc/osapi/pki`\n" +
		"by default). Give the fingerprint to `osapi client admission admit`. The\n" +
		"controller admits the machine only when the fingerprint is the one of the\n" +
		"key in its request. If no key exists, the command makes a new key pair.",
	Example: `  osapi agent admission`,
	Run: func(_ *cobra.Command, _ []string) {
		keyDir := appConfig.Agent.PKI.KeyDir
		if keyDir == "" {
			keyDir = "/etc/osapi/pki"
		}

		fs := osfs.NewWithNoIdm()
		mgr := pki.New(fs, keyDir, "agent")
		if err := mgr.LoadOrGenerate(); err != nil {
			cli.HandleError(fmt.Errorf("load agent key from %s: %w", keyDir, err), logger)

			return
		}

		ident, err := identity.GetIdentity(fs, appConfig.Agent.Hostname)
		if err != nil {
			cli.HandleError(err, logger)

			return
		}

		if jsonOutput {
			fmt.Printf("{\"machine_id\":%q,\"hostname\":%q,\"fingerprint\":%q}\n",
				ident.MachineID, ident.Hostname, mgr.Fingerprint())

			return
		}

		cli.PrintHeading("Admission")
		cli.PrintKV(
			"Machine ID", ident.MachineID,
			"Hostname", ident.Hostname,
			"Fingerprint", mgr.Fingerprint(),
		)
		cli.PrintNote(fmt.Sprintf(
			"To admit: osapi client admission admit --machine-id %s --fingerprint %s",
			ident.MachineID, mgr.Fingerprint()))
	},
}

func init() {
	agentCmd.AddCommand(agentAdmissionCmd)
}
