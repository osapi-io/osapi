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

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// clientFileUploadCmd represents the clientFileUpload command.
var clientFileUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to the Object Store",
	Long: "An upload of the same content changes nothing (`Changed: false`). If the\n" +
		"file exists with different content, the command rejects the upload with 409\n" +
		"Conflict. `file update` replaces the file.",
	Example: `  osapi client file upload --name app.conf --file /tmp/app.conf
  osapi client file upload --name app.conf.tmpl --file /tmp/app.conf.tmpl \
    --content-type template`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		name, _ := cmd.Flags().GetString("name")
		filePath, _ := cmd.Flags().GetString("file")
		contentType, _ := cmd.Flags().GetString("content-type")

		f, err := os.Open(filePath)
		if err != nil {
			cli.LogFatal(logger, "failed to open file", err)
		}
		defer func() { _ = f.Close() }()

		resp, err := sdkClient.File.Upload(ctx, name, contentType, f)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		printFileWrite(resp)
	},
}

func init() {
	clientFileCmd.AddCommand(clientFileUploadCmd)

	clientFileUploadCmd.PersistentFlags().
		String("name", "", "Name for the file in the Object Store")
	clientFileUploadCmd.PersistentFlags().
		String("file", "", "Path to the local file to upload")
	clientFileUploadCmd.PersistentFlags().
		String("content-type", "raw", "File type: raw or template (default raw)")

	_ = clientFileUploadCmd.MarkPersistentFlagRequired("name")
	_ = clientFileUploadCmd.MarkPersistentFlagRequired("file")
}

// printFileWrite prints the file that an upload or an update wrote: the
// answer as it came with --json.
func printFileWrite(
	resp *client.Response[client.FileUpload],
) {
	if jsonOutput {
		rawJSON := resp.RawJSON()
		if rawJSON == nil {
			rawJSON, _ = json.Marshal(resp.Data)
		}
		cli.PrintJSON(rawJSON)

		return
	}

	fmt.Println()
	cli.PrintKV("Name", resp.Data.Name)
	cli.PrintKV("SHA256", resp.Data.SHA256)
	cli.PrintKV("Size", fmt.Sprintf("%d", resp.Data.Size))
	cli.PrintKV("Changed", fmt.Sprintf("%v", resp.Data.Changed))
	cli.PrintKV("Content-Type", resp.Data.ContentType)
}
