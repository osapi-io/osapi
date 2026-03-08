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
	"strings"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientNodeFileDeployCmd represents the clientNodeFileDeploy command.
var clientNodeFileDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a file from Object Store to a host",
	Long: `Deploy a file from the OSAPI Object Store to the target host's filesystem.
The file is fetched from the Object Store and written to the specified path.
SHA-256 idempotency ensures unchanged files are not rewritten.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		objectName, _ := cmd.Flags().GetString("object")
		path, _ := cmd.Flags().GetString("path")
		contentType, _ := cmd.Flags().GetString("content-type")
		mode, _ := cmd.Flags().GetString("mode")
		owner, _ := cmd.Flags().GetString("owner")
		group, _ := cmd.Flags().GetString("group")
		varFlags, _ := cmd.Flags().GetStringSlice("var")

		if host == "_all" {
			fmt.Print("This will deploy the file to ALL hosts. Continue? [y/N] ")
			var confirm string
			if _, err := fmt.Scanln(&confirm); err != nil || (confirm != "y" && confirm != "Y") {
				fmt.Println("Aborted.")
				return
			}
		}

		vars := parseVarFlags(varFlags)

		resp, err := sdkClient.Node.FileDeploy(ctx, client.FileDeployOpts{
			Target:      host,
			ObjectName:  objectName,
			Path:        path,
			ContentType: contentType,
			Mode:        mode,
			Owner:       owner,
			Group:       group,
			Vars:        vars,
		})
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		if resp.Data.JobID != "" {
			fmt.Println()
			cli.PrintKV("Job ID", resp.Data.JobID)
		}

		changed := resp.Data.Changed
		changedPtr := &changed
		results := []cli.MutationResultRow{
			{
				Hostname: resp.Data.Hostname,
				Changed:  changedPtr,
			},
		}
		headers, rows := cli.BuildMutationTable(results, nil)
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

// parseVarFlags converts a slice of "key=value" strings into a map.
func parseVarFlags(
	flags []string,
) map[string]interface{} {
	if len(flags) == 0 {
		return nil
	}

	vars := make(map[string]interface{}, len(flags))
	for _, f := range flags {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}

	return vars
}

func init() {
	clientNodeFileCmd.AddCommand(clientNodeFileDeployCmd)

	clientNodeFileDeployCmd.PersistentFlags().
		String("object", "", "Name of the file in the Object Store (required)")
	clientNodeFileDeployCmd.PersistentFlags().
		String("path", "", "Destination path on the target filesystem (required)")
	clientNodeFileDeployCmd.PersistentFlags().
		String("content-type", "raw", "Content type: raw or template (default raw)")
	clientNodeFileDeployCmd.PersistentFlags().
		String("mode", "", "File permission mode (e.g., 0644)")
	clientNodeFileDeployCmd.PersistentFlags().
		String("owner", "", "File owner user")
	clientNodeFileDeployCmd.PersistentFlags().
		String("group", "", "File owner group")
	clientNodeFileDeployCmd.PersistentFlags().
		StringSlice("var", []string{}, "Template variable as key=value (repeatable)")

	_ = clientNodeFileDeployCmd.MarkPersistentFlagRequired("object")
	_ = clientNodeFileDeployCmd.MarkPersistentFlagRequired("path")
}
