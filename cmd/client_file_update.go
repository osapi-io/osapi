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
	"os"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/reference"
)

// clientFileUpdateCmd represents the file update command.
var clientFileUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Replace a file in the Object Store",
	Long: "The command replaces the content of a stored file, or stores it when the\n" +
		"name is new. The same content changes nothing (`Changed: false`).",
	Annotations: map[string]string{
		reference.Permission: "file:write",
	},
	Example: `  osapi client file update --name app.conf --file /tmp/app.conf`,
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

		resp, err := sdkClient.File.Update(ctx, name, contentType, f)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		printFileWrite(resp)
	},
}

func init() {
	clientFileCmd.AddCommand(clientFileUpdateCmd)

	clientFileUpdateCmd.PersistentFlags().
		String("name", "", "Name of the file in the Object Store")
	clientFileUpdateCmd.PersistentFlags().
		String("file", "", "Path to the local file with the new content")
	clientFileUpdateCmd.PersistentFlags().
		String("content-type", "raw", "File type: raw or template")

	_ = clientFileUpdateCmd.MarkPersistentFlagRequired("name")
	_ = clientFileUpdateCmd.MarkPersistentFlagRequired("file")
}
