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

package behaviorcmd

import (
	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli"
)

// Deps is what a generated command needs from cmd. Each is a function
// because the SDK client and the --json flag are only known when the command
// runs, not when it is built.
type Deps struct {
	// Caller returns the SDK client.
	Caller func() Caller
	// JSON reports whether --json was given.
	JSON func() bool
	// Error reports a failed call.
	Error func(err error)
}

type entry struct {
	group string
	cmd   *cobra.Command
}

var (
	registered []entry
	deps       Deps
)

// Register is called by the init of each behavior's generated gen/cli.go
// with the behavior's command, its subcommands already added.
func Register(
	group string,
	cmd *cobra.Command,
) {
	registered = append(registered, entry{group: group, cmd: cmd})
}

// Attach adds every registered behavior command under parent and keeps d for
// the commands to call through. A behavior with a group is nested under a
// command named for the group.
func Attach(
	parent *cobra.Command,
	d Deps,
) {
	deps = d

	groups := map[string]*cobra.Command{}
	for _, e := range registered {
		into := parent
		if e.group != "" {
			g, ok := groups[e.group]
			if !ok {
				g = &cobra.Command{Use: e.group, Short: "The " + e.group + " behaviors"}
				parent.AddCommand(g)
				groups[e.group] = g
			}
			into = g
		}
		into.AddCommand(e.cmd)
	}
}

// Run calls one method with in and prints the result: the API body for
// --json, otherwise the job ID and the table every behavior renders. Every
// generated command's Run calls it.
func Run(
	cmd *cobra.Command,
	b *behavior.Behavior,
	method string,
	in any,
) {
	d := deps
	m := b.Method(method)
	target, _ := cmd.Flags().GetString("target")

	result, raw, err := Call(cmd.Context(), d.Caller(), b, m, target, in)
	if err != nil {
		d.Error(err)

		return
	}

	if d.JSON() {
		cli.PrintJSON(raw)

		return
	}

	cli.PrintJobID(result.JobID)
	cli.PrintCompactTable([]cli.Section{Table(m, result)})
}
