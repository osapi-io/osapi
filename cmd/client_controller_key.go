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

	"github.com/retr0h/osapi/internal/agent/pki"
	"github.com/retr0h/osapi/internal/cli"
)

// clientControllerKeyCmd represents the controller key parent command.
var clientControllerKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Controller key management",
}

// clientControllerKeyFingerprintCmd prints the local controller's key fingerprint.
var clientControllerKeyFingerprintCmd = &cobra.Command{
	Use:   "fingerprint",
	Short: "Show the local controller key fingerprint",
	Long:  `Read the local controller key and display its SHA256 fingerprint.`,
	Run: func(_ *cobra.Command, _ []string) {
		keyDir := appConfig.Controller.PKI.KeyDir
		if keyDir == "" {
			keyDir = "/etc/osapi/pki"
		}

		fs := osfs.NewWithNoIdm()
		mgr := pki.NewManager(fs, keyDir)

		if err := mgr.LoadOrGenerate(); err != nil {
			cli.HandleError(
				fmt.Errorf("load controller key from %s: %w", keyDir, err),
				logger,
			)
			return
		}

		fingerprint := mgr.Fingerprint()
		if fingerprint == "" {
			fmt.Println("No controller key found.")
			return
		}

		if jsonOutput {
			fmt.Printf(`{"fingerprint":"%s"}`, fingerprint)
			fmt.Println()
			return
		}

		fmt.Println()
		cli.PrintKV("Fingerprint", fingerprint)
	},
}

func init() {
	clientControllerCmd.AddCommand(clientControllerKeyCmd)
	clientControllerKeyCmd.AddCommand(clientControllerKeyFingerprintCmd)
}
