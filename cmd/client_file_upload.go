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
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// clientFileUploadCmd represents the clientFileUpload command.
var clientFileUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to the Object Store",
	Long:  `Upload a local file to the OSAPI Object Store for later deployment.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		name, _ := cmd.Flags().GetString("name")
		filePath, _ := cmd.Flags().GetString("file")
		contentType, _ := cmd.Flags().GetString("content-type")
		force, _ := cmd.Flags().GetBool("force")

		f, err := os.Open(filePath)
		if err != nil {
			cli.LogFatal(logger, "failed to open file", err)
		}
		defer func() { _ = f.Close() }()

		var opts []client.UploadOption
		if force {
			opts = append(opts, client.WithForce())
		}

		resp, err := sdkClient.File.Upload(ctx, name, contentType, f, opts...)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			rawJSON := resp.RawJSON()
			if rawJSON == nil {
				out, _ := json.Marshal(resp.Data)
				rawJSON = out
			}
			fmt.Println(string(rawJSON))
			return
		}

		fmt.Println()
		cli.PrintKV("Name", resp.Data.Name)
		cli.PrintKV("SHA256", resp.Data.SHA256)
		cli.PrintKV("Size", fmt.Sprintf("%d", resp.Data.Size))
		cli.PrintKV("Changed", fmt.Sprintf("%v", resp.Data.Changed))
		cli.PrintKV("Content-Type", resp.Data.ContentType)
	},
}

func init() {
	clientFileCmd.AddCommand(clientFileUploadCmd)

	clientFileUploadCmd.PersistentFlags().
		String("name", "", "Name for the file in the Object Store (required)")
	clientFileUploadCmd.PersistentFlags().
		String("file", "", "Path to the local file to upload (required)")
	clientFileUploadCmd.PersistentFlags().
		String("content-type", "raw", "File type: raw or template (default raw)")
	clientFileUploadCmd.PersistentFlags().
		Bool("force", false, "Force upload even if file already exists with different content")

	_ = clientFileUploadCmd.MarkPersistentFlagRequired("name")
	_ = clientFileUploadCmd.MarkPersistentFlagRequired("file")
}
