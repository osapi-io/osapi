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
	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/reference"
	"github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/internal/controller/endpoints/facts"
)

var clientFactsCmd = &cobra.Command{
	Use:   "facts",
	Short: "Read the facts that an input can refer to",
	Annotations: map[string]string{
		reference.Permission: "agent:read",
	},
}

var clientFactsKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Read the keys of @fact references",
	Annotations: map[string]string{
		reference.Permission: "agent:read",
	},
}

func init() {
	clientCmd.AddCommand(clientFactsCmd)
	clientFactsCmd.AddCommand(clientFactsKeysCmd)

	// osapi client facts keys list is generated from the facts group's
	// definition; this is how its answer prints.
	command.Render(facts.Group, "list", renderFactKeys)
}

// renderFactKeys prints each fact key with its description.
func renderFactKeys(
	_ *cobra.Command,
	_ facts.KeysInput,
	out facts.FactKeysResponse,
) {
	rows := make([][]string, 0, len(out.Items))
	for _, k := range out.Items {
		rows = append(rows, []string{"@fact." + k.Key, k.Description})
	}

	cli.PrintCompactTable([]cli.Section{
		{
			Title:   "Fact keys",
			Empty:   "None.",
			Headers: []string{"REFERENCE", "DESCRIPTION"},
			Rows:    rows,
		},
	})
}
