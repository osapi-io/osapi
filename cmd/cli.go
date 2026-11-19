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
)

// cliCmd is the parent of the commands about the CLI itself.
var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Describe and check the commands of this binary",
}

// cliCommandCmd is the parent of the commands of the command DSL.
var cliCommandCmd = &cobra.Command{
	Use:   "command",
	Short: "List, describe and validate the local commands, and print their DSL",
	Long: "A local command runs in the CLI process and calls no API, such as\n" +
		"`osapi token generate`. Its group is a definition in\n" +
		"`internal/cli/commands/<group>`, and its commands, flags, help and pages\n" +
		"come from it. A command that calls the API is an endpoint\n" +
		"(`osapi controller endpoint`) or a behavior (`osapi behavior`). These\n" +
		"commands read the definitions in the binary. Thus they need no config file\n" +
		"and no running controller.",
}

func init() {
	rootCmd.AddCommand(cliCmd)
	cliCmd.AddCommand(cliCommandCmd)
	skipConfigCmds["cli"] = true
}
