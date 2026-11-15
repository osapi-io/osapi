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

package command

import (
	"fmt"

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
	next  bool
	cmd   *cobra.Command
}

var (
	registered []entry
	deps       Deps
)

// Register is called by the init of each behavior's generated gen/cli.go
// with the behavior's command, its subcommands already added, and whether
// the behavior is Next.
func Register(
	group string,
	next bool,
	cmd *cobra.Command,
) {
	registered = append(registered, entry{group: group, next: next, cmd: cmd})
}

// HasNext reports whether any registered behavior is Next, so osapi client
// next is shown only while one is.
func HasNext() bool {
	for _, e := range registered {
		if e.next {
			return true
		}
	}

	return false
}

// Attach adds every registered behavior command under node, or under next
// for a Next behavior, and keeps d for the commands to call through. A
// behavior with a group is nested under a command named for the group: the
// parent's own command of that name if it has one, such as a hand-written
// network, or a new one.
//
// A behavior command whose name its parent already has is an error: a
// hand-written domain the behavior replaces is still there. That is what a
// cutover leaves until the hand-written commands are deleted, so the error
// says which.
func Attach(
	node *cobra.Command,
	next *cobra.Command,
	d Deps,
) error {
	deps = d

	for _, e := range registered {
		into := node
		if e.next {
			into = next
		}
		if e.group != "" {
			into = group(into, e.group)
		}
		for _, c := range into.Commands() {
			if c.Name() == e.cmd.Name() {
				return fmt.Errorf(
					"%s %s is both a hand-written command and a behavior's: "+
						"delete the hand-written one from cmd/, or mark the behavior behavior.Next "+
						"until its cutover",
					into.CommandPath(), e.cmd.Name(),
				)
			}
		}
		into.AddCommand(e.cmd)
	}

	return nil
}

// group is parent's command named name, added if parent has none.
func group(
	parent *cobra.Command,
	name string,
) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	g := &cobra.Command{Use: name, Short: "The " + name + " behaviors"}
	parent.AddCommand(g)

	return g
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
