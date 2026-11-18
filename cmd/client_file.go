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

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/reference"
	"github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/internal/controller/endpoints/file"
)

// clientFileCmd represents the clientFile command.
var clientFileCmd = &cobra.Command{
	Use:   "file",
	Short: "Manage the files in the Object Store",
}

// clientFileStaleCmd holds the commands of the stale deployments, such as
// osapi client file stale list.
var clientFileStaleCmd = &cobra.Command{
	Use:   "stale",
	Short: "Read the deployments whose file changed since",
	Annotations: map[string]string{
		reference.Permission: "file:read",
	},
}

func init() {
	clientCmd.AddCommand(clientFileCmd)
	clientFileCmd.AddCommand(clientFileStaleCmd)

	// The file group's commands are generated from its definition; this is
	// how each answer prints.
	command.Render(file.Group, "list", renderFileList)
	command.Render(file.Group, "get", renderFileGet)
	command.Render(file.Group, "upload", renderFileWrite[file.UploadInput])
	command.Render(file.Group, "update", renderFileWrite[file.UpdateInput])
	command.Acts(file.Group, "delete", renderFileDeleted)
	command.Render(file.Group, "stale", renderFileStale)
}

// renderFileList prints each stored file.
func renderFileList(
	_ *cobra.Command,
	_ file.ListInput,
	out file.FileListResponse,
) {
	if len(out.Items) == 0 {
		cli.PrintHeading("Files (0)")
		cli.PrintNote("None. Upload one with osapi client file upload.")

		return
	}

	rows := make([][]string, 0, len(out.Items))
	for _, f := range out.Items {
		rows = append(rows, []string{f.Name, f.Sha256, fmt.Sprintf("%d", f.Size), f.Source})
	}
	cli.PrintCompactTable([]cli.Section{{
		Title:   fmt.Sprintf("Files (%d)", out.TotalItems),
		Headers: []string{"NAME", "SHA256", "SIZE", "SOURCE"},
		Rows:    rows,
	}})
}

// renderFileGet prints the metadata of a file.
func renderFileGet(
	_ *cobra.Command,
	_ file.GetInput,
	out file.FileInfoResponse,
) {
	fmt.Println()
	cli.PrintKV("Name", out.Name)
	cli.PrintKV("SHA256", out.Sha256)
	cli.PrintKV("Size", fmt.Sprintf("%d", out.Size))
}

// renderFileWrite prints the file that an upload or an update wrote.
func renderFileWrite[In any](
	_ *cobra.Command,
	_ In,
	out file.FileUploadResponse,
) {
	fmt.Println()
	cli.PrintKV("Name", out.Name)
	cli.PrintKV("SHA256", out.Sha256)
	cli.PrintKV("Size", fmt.Sprintf("%d", out.Size))
	cli.PrintKV("Changed", fmt.Sprintf("%v", out.Changed))
	cli.PrintKV("Content-Type", out.ContentType)
}

// renderFileDeleted prints the file that went. A delete answers with no
// body, so --json prints what it did.
func renderFileDeleted(
	_ *cobra.Command,
	in file.DeleteInput,
	_ struct{},
) {
	if jsonOutput {
		result, _ := json.Marshal(map[string]string{"name": in.Name, "status": "deleted"})
		cli.PrintJSON(result)

		return
	}

	cli.PrintHeading("File")
	cli.PrintKV("Name", in.Name, "Status", "Deleted")
}

// renderFileStale prints each deployment whose file changed since.
func renderFileStale(
	_ *cobra.Command,
	_ file.StaleInput,
	out file.StaleDeploymentsResponse,
) {
	if len(out.Items) == 0 {
		cli.PrintHeading("Stale deployments (0)")
		cli.PrintNote("None. Every deployment is in sync.")

		return
	}

	short := func(sha string) string {
		if len(sha) > 12 {
			return sha[:12] + "…"
		}

		return sha
	}
	rows := make([][]string, 0, len(out.Items))
	for _, s := range out.Items {
		current := short(s.CurrentSha)
		if s.CurrentSha == "" {
			current = "(deleted)"
		}
		rows = append(rows, []string{
			s.ObjectName, s.Hostname, s.Provider, s.DeployedAt, short(s.DeployedSha), current,
		})
	}
	cli.PrintCompactTable([]cli.Section{{
		Title: fmt.Sprintf("Stale deployments (%d)", out.TotalItems),
		Headers: []string{
			"OBJECT", "HOSTNAME", "PROVIDER", "DEPLOYED", "DEPLOYED SHA", "CURRENT SHA",
		},
		Rows: rows,
	}})
}
