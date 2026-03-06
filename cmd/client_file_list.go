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
)

// clientFileListCmd represents the clientFileList command.
var clientFileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored files",
	Long:  `List all files stored in the OSAPI Object Store.`,
	Run: func(cmd *cobra.Command, _ []string) {
		// TODO(sdk): Replace with SDK call when FileService is available:
		//   ctx := cmd.Context()
		//   resp, err := sdkClient.File.List(ctx)
		//   if err != nil {
		//       cli.HandleError(err, logger)
		//       return
		//   }
		//
		//   if jsonOutput {
		//       fmt.Println(string(resp.RawJSON()))
		//       return
		//   }
		//
		//   files := resp.Data.Files
		//   if len(files) == 0 {
		//       fmt.Println("No files found.")
		//       return
		//   }
		//
		//   rows := make([][]string, 0, len(files))
		//   for _, f := range files {
		//       rows = append(rows, []string{
		//           f.Name,
		//           f.SHA256,
		//           fmt.Sprintf("%d", f.Size),
		//       })
		//   }
		//
		//   sections := []cli.Section{
		//       {
		//           Title:   fmt.Sprintf("Files (%d)", resp.Data.Total),
		//           Headers: []string{"NAME", "SHA256", "SIZE"},
		//           Rows:    rows,
		//       },
		//   }
		//   cli.PrintCompactTable(sections)

		_ = cmd.Context()
		logger.Error("file list requires osapi-sdk FileService (not yet available)")
		fmt.Println("file list: SDK FileService not yet integrated")
	},
}

func init() {
	clientFileCmd.AddCommand(clientFileListCmd)
}
