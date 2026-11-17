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
	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// endpointMethod is one method in `controller endpoint get --json`.
type endpointMethod struct {
	Name        string `json:"name"`
	Verb        string `json:"verb"`
	Path        string `json:"path"`
	Permission  string `json:"permission"`
	OperationID string `json:"operation_id"`
}

var endpointGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Describe one endpoint group: its routes, permissions and operation IDs",
	Long: "`--openapi` prints the group's part of the spec, as the definition makes it.\n" +
		"Read it beside the hand-written spec of the group to compare the prose.\n" +
		"`TestEndpointParity` compares the rest.",
	Example: `  osapi controller endpoint get --name agent
  osapi controller endpoint get --name agent --json
  osapi controller endpoint get --name agent --openapi`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		name, _ := cmd.Flags().GetString("name")
		openapi, _ := cmd.Flags().GetBool("openapi")

		g := endpoint.Lookup(name)
		if g == nil {
			cli.LogFatal(logger, "no endpoint group", fmt.Errorf("no group named %q", name))
		}

		if openapi {
			doc, err := endpoint.OpenAPI(g, endpoint.ErrorRef)
			if err != nil {
				cli.LogFatal(logger, "spec does not build", err)
			}
			raw, _ := json.Marshal(doc)
			var tree any
			_ = json.Unmarshal(raw, &tree)
			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			_ = enc.Encode(tree)

			return
		}

		methods := make([]endpointMethod, 0, len(g.Methods))
		for _, m := range g.Methods {
			perm := g.Permission(m)
			if perm == "" {
				perm = "none (public)"
			}
			methods = append(methods, endpointMethod{
				Name: m.Name, Verb: m.Verb, Path: g.Path(m),
				Permission: perm, OperationID: m.OperationID(g),
			})
		}

		if jsonOutput {
			out, _ := json.Marshal(map[string]any{"name": g.Name, "doc": g.Doc, "methods": methods})
			cli.PrintJSON(out)

			return
		}

		cli.PrintHeading("Endpoint group")
		cli.PrintKV("Name", g.Name, "Description", g.Doc)
		rows := make([][]string, 0, len(methods))
		for _, m := range methods {
			rows = append(rows, []string{m.Name, m.Verb, m.Path, m.Permission, m.OperationID})
		}
		cli.PrintCompactTable([]cli.Section{{
			Title:     fmt.Sprintf("Methods (%d)", len(methods)),
			Headers:   []string{"NAME", "VERB", "PATH", "PERMISSION", "OPERATION ID"},
			Rows:      rows,
			Reference: true,
		}})
	},
}

func init() {
	endpointCmd.AddCommand(endpointGetCmd)
	endpointGetCmd.Flags().String("name", "", "The endpoint group, such as agent (required)")
	endpointGetCmd.Flags().
		Bool("openapi", false, "Print the group's part of the OpenAPI spec as YAML")
	_ = endpointGetCmd.MarkFlagRequired("name")
}
