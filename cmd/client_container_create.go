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
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// clientContainerCreateCmd represents the clientContainerCreate command.
var clientContainerCreateCmd = &cobra.Command{
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

		body := gen.ContainerCreateRequest{
			Image:     image,
			AutoStart: &autoStart,
		}

		if name != "" {
			body.Name = &name
		}
		if len(envFlags) > 0 {
			body.Env = &envFlags
		}
		if len(portFlags) > 0 {
			body.Ports = &portFlags
		}
		if len(volumeFlags) > 0 {
			body.Volumes = &volumeFlags
		}

		resp, err := sdkClient.Container.Create(ctx, host, body)
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

		for _, r := range resp.Data.Results {
			cli.PrintKV("Hostname", r.Hostname)
			if r.Error != "" {
				cli.PrintKV("Error", r.Error)
				continue
			}
			cli.PrintKV(
				"Container ID", r.ID,
				"Name", r.Name,
			)
			cli.PrintKV(
				"Image", r.Image,
				"State", r.State,
			)
		}
	},
}

func init() {
	clientContainerCmd.AddCommand(clientContainerCreateCmd)

	clientContainerCreateCmd.PersistentFlags().
		String("image", "", "Container image reference (required)")
	clientContainerCreateCmd.PersistentFlags().
		String("name", "", "Optional name for the container")
	clientContainerCreateCmd.PersistentFlags().
		StringSlice("env", []string{}, "Environment variable in KEY=VALUE format (repeatable)")
	clientContainerCreateCmd.PersistentFlags().
		StringSlice("port", []string{}, "Port mapping in host:container format (repeatable)")
	clientContainerCreateCmd.PersistentFlags().
		StringSlice("volume", []string{}, "Volume mount in host:container format (repeatable)")
	clientContainerCreateCmd.PersistentFlags().
		Bool("auto-start", true, "Start the container immediately after creation")

	_ = clientContainerCreateCmd.MarkPersistentFlagRequired("image")
}
