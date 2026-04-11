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

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// clientContainerDockerCreateCmd represents the clientContainerDockerCreate command.
var clientContainerDockerCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new container",
	Long:  `Create a new container on the target node from the specified image.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		image, _ := cmd.Flags().GetString("image")
		name, _ := cmd.Flags().GetString("name")
		envFlags, _ := cmd.Flags().GetStringSlice("env")
		portFlags, _ := cmd.Flags().GetStringSlice("port")
		volumeFlags, _ := cmd.Flags().GetStringSlice("volume")
		autoStart, _ := cmd.Flags().GetBool("auto-start")

		opts := client.DockerCreateOpts{
			Image:     image,
			Name:      name,
			AutoStart: &autoStart,
			Env:       envFlags,
			Ports:     portFlags,
			Volumes:   volumeFlags,
		}

		resp, err := sdkClient.Docker.Create(ctx, host, opts)
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

		results := make([]cli.ResultRow, 0)
		for _, r := range resp.Data.Results {
			var errPtr *string
			if r.Error != "" {
				e := r.Error
				errPtr = &e
			}
			changed := r.Changed
			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Status:   r.Status,
				Changed:  &changed,
				Error:    errPtr,
				Fields: []string{
					r.ID,
					r.Name,
					r.Image,
					r.State,
				},
			})
		}
		tr := cli.BuildMutationTable(
			results,
			[]string{"ID", "NAME", "IMAGE", "STATE"},
		)
		cli.PrintCompactTable([]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}})
	},
}

func init() {
	clientContainerDockerCmd.AddCommand(clientContainerDockerCreateCmd)

	clientContainerDockerCreateCmd.PersistentFlags().
		String("image", "", "Container image reference (required)")
	clientContainerDockerCreateCmd.PersistentFlags().
		String("name", "", "Optional name for the container")
	clientContainerDockerCreateCmd.PersistentFlags().
		StringSlice("env", []string{}, "Environment variable in KEY=VALUE format (repeatable)")
	clientContainerDockerCreateCmd.PersistentFlags().
		StringSlice("port", []string{}, "Port mapping in host:container format (repeatable)")
	clientContainerDockerCreateCmd.PersistentFlags().
		StringSlice("volume", []string{}, "Volume mount in host:container format (repeatable)")
	clientContainerDockerCreateCmd.PersistentFlags().
		Bool("auto-start", true, "Start the container immediately after creation")

	_ = clientContainerDockerCreateCmd.MarkPersistentFlagRequired("image")
}
