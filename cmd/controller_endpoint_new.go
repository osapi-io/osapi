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

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

var endpointNewCmd = &cobra.Command{
	Use:   "create",
	Short: "Write a new endpoint group that builds and passes its tests",
	Long: "It writes `internal/controller/endpoints/<name>/` in the osapi checkout in\n" +
		"the current directory, and adds its import to\n" +
		"`internal/controller/endpoints/all`. The group has a list and a get, an empty\n" +
		"`Deps` and the tests of its methods. Then run `go generate\n" +
		"./internal/controller/endpoints/all` and `just generate`. While its `Deps` is\n" +
		"empty, the controller needs no wiring. `osapi controller endpoint spec get` lists\n" +
		"each part of the DSL.",
	Example: `  osapi controller endpoint create --name trace --description "The traces of the controller"`,
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		name, _ := cmd.Flags().GetString("name")
		doc, _ := cmd.Flags().GetString("description")

		files, err := endpoint.Scaffold(".", name, doc)
		if err != nil {
			cli.LogFatal(logger, "endpoint new", err)
		}

		cli.PrintHeading(fmt.Sprintf("Wrote (%d)", len(files)))
		for _, f := range files {
			fmt.Println("  " + f)
		}
		cli.PrintHeading("Next")
		for _, step := range []string{
			"go generate ./internal/controller/endpoints/all",
			"mise exec -- just generate",
			"go test ./internal/controller/endpoints/" + name + "/...",
			"go run . controller endpoint validate",
		} {
			fmt.Println("  " + step)
		}
	},
}

func init() {
	endpointCmd.AddCommand(endpointNewCmd)
	endpointNewCmd.Flags().
		String("name", "", "The group's name: its path segment and CLI group")
	endpointNewCmd.Flags().String("description", "", "What the group is, in one line")
	_ = endpointNewCmd.MarkFlagRequired("name")
	_ = endpointNewCmd.MarkFlagRequired("description")
}
