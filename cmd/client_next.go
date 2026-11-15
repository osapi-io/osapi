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
	"os"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/behavior/command"
	"github.com/osapi-io/osapi/internal/cli"

	// Every behavior with a CLI registers its generated commands.
	_ "github.com/osapi-io/osapi/internal/behaviors/all/gen"
)

// clientNextCmd holds the commands of behaviors marked Next. They call the
// /api/next routes and sit beside the hand-written commands of the domains
// they replace, until each domain's cutover. Every other behavior's commands
// are under osapi client node. It is hidden while no behavior is Next.
var clientNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Commands of behaviors served under /api/next",
}

var clientNextNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Behaviors that run on a node",
}

func init() {
	clientCmd.AddCommand(clientNextCmd)
	clientNextCmd.AddCommand(clientNextNodeCmd)

	clientNextNodeCmd.PersistentFlags().
		StringP("target", "T", "_all", "Target: _all, _any, hostname, or label (group:web.dev)")

	clientNextCmd.Hidden = !command.HasNext()

	if err := command.Attach(clientNodeCmd, clientNextNodeCmd, command.Deps{
		Caller: func() command.Caller { return sdkClient },
		JSON:   func() bool { return jsonOutput },
		Error:  func(err error) { cli.HandleError(err, logger) },
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
