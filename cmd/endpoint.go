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

	// Register every endpoint group, so the endpoint commands see them.
	_ "github.com/osapi-io/osapi/internal/endpoints/all"
)

var endpointCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "List, describe and validate the controller's own endpoints",
	Long: "An endpoint is one operation of the controller itself, such as\n" +
		"`GET /api/agent`. It runs on the controller, not on a host. Work on a host\n" +
		"is a behavior: see `osapi behavior`. These commands read the definitions in\n" +
		"the binary. Thus they need no config file and no controller.\n\n" +
		"The groups are defined, and their specs are compared with the hand-written\n" +
		"specs. The controller still serves the hand-written specs.",
}

func init() {
	rootCmd.AddCommand(endpointCmd)
	skipConfigCmds["endpoint"] = true
}
