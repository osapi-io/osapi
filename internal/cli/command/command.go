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

// Package command is the command layer of the CLI: the commands of a group
// under osapi client, with no HTTP. A group's gen/cli.go makes one cobra
// command for each method and registers it here; Attach puts them under
// osapi client <group>. The layer that sends a command's request, such as
// the endpoint layer, gives the answer to Output, which prints it: the body
// for --json, or the group's renderer for the method.
package command

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli/reference"
)

// Group is a group of commands, osapi client <group>, with one command for
// each of its methods. An endpoint group is one.
type Group interface {
	// GroupName is the group's command, and the key of its renderers.
	GroupName() string
	// GroupDoc says what the group is, in one line: its command's help.
	GroupDoc() string
	// GroupNotes is the Notes section of the page of the group's command,
	// or "".
	GroupNotes() string
	// CommandPath is the words of the group's command under osapi client,
	// such as next node sysctl.
	CommandPath() []string
	// Commands are the group's methods, in the order they were declared.
	Commands() []Method
	// CommandPermission is the permission the command of method m needs, or
	// "" for one that needs none.
	CommandPermission(m Method) string
}

// Method is one command of a group.
type Method interface {
	// MethodName is the method's name in its group, the key of its renderer.
	MethodName() string
	// Input is the type of the method's input: a flag for each field.
	Input() reflect.Type
	// CommandName is the command's name.
	CommandName() string
	// SubPath is the command the command goes under in its group, or "".
	SubPath() string
	// CommandShort is the command's one line of help.
	CommandShort() string
	// CommandLong is the rest of the command's help, or "".
	CommandLong() string
	// Examples are the command's examples.
	Examples() []Example
	// CommandFlags are the command's flags that are not part of the input.
	CommandFlags() []Flag
	// PageNotes is the Notes section of the command's page, or "".
	PageNotes() string
	// PageResult is the Result section of the command's page, or "".
	PageResult() string
}

// Example is an example of a CLI command: its flags, and what it shows,
// which the help writes as a # line above it.
type Example struct {
	Flags string
	Doc   string
}

// Flag is a flag of the CLI command that is not part of the request: it
// changes only how the answer prints, such as --all. The renderer reads it.
type Flag struct {
	// Name is the flag's name, without the dashes.
	Name string
	// Usage is its help.
	Usage string
	// String is true for a string flag, which StringFlag adds; false for a
	// bool flag.
	String bool
	// Value is a string flag's value when the caller gives none.
	Value string
	// Required is true for a string flag the caller must give.
	Required bool
}

// Deps is what the commands need from the CLI.
type Deps struct {
	// JSON reports whether --json was given.
	JSON func() bool
	// Error reports a failed call.
	Error func(err error)
	// PrintJSON prints an answer as it came, for --json.
	PrintJSON func(body []byte)
}

type entry struct {
	group Group
	cmds  []*cobra.Command
}

var (
	registered []entry
	deps       Deps
	renderers  = map[string]renderer{}
)

// renderer prints a method's answer: fn gets the input and a value of out,
// the type the answer decodes into.
type renderer struct {
	fn  func(cmd *cobra.Command, in, out any)
	out reflect.Type
	// acts is true for a renderer that Acts set, which runs with --json too.
	acts bool
}

// SubKey is the annotation of a generated command that goes under a
// command of its own: the Sub segment of its method, such as trace in
// osapi client telemetry trace get.
const SubKey = "osapi.endpoint.sub"

// Register adds a group's generated commands. Each gen/cli.go calls it from
// init.
func Register(
	g Group,
	cmds ...*cobra.Command,
) {
	registered = append(registered, entry{group: g, cmds: cmds})
}

// Attach adds every registered group's commands under parent, osapi client,
// and keeps d for the commands to call through. A group goes under the
// commands of its command path: next node sysctl is under next, then node,
// which Attach adds when parent does not have them. A group or a command
// that is among groups, by name, is that one, so the generated commands
// join a hand-written command of the same name, such as agent with key and
// status.
func Attach(
	parent *cobra.Command,
	d Deps,
	groups ...*cobra.Command,
) error {
	deps = d
	for _, e := range registered {
		name := e.group.GroupName()
		words := e.group.CommandPath()
		into := parent
		for _, w := range words[:len(words)-1] {
			next := child(into, w)
			if next == nil {
				next = &cobra.Command{Use: w, Short: strings.ToUpper(w[:1]) + w[1:] + " commands"}
				into.AddCommand(next)
			}
			into = next
		}
		group := child(into, name)
		for _, g := range groups {
			if g.Name() == name {
				group = g
			}
		}
		if group == nil {
			group = &cobra.Command{Use: name, Short: e.group.GroupDoc()}
			if notes := e.group.GroupNotes(); notes != "" {
				group.Annotations = map[string]string{reference.Notes: notes}
			}
			into.AddCommand(group)
		}
		for _, c := range e.cmds {
			parent := group
			if sub := c.Annotations[SubKey]; sub != "" {
				parent = child(group, sub)
				for _, g := range groups {
					if g.Name() == sub && g != group {
						parent = g
					}
				}
				if parent == nil {
					parent = &cobra.Command{
						Use:   sub,
						Short: strings.ToUpper(sub[:1]) + sub[1:] + " commands",
					}
					group.AddCommand(parent)
				}
			}
			if child(parent, c.Name()) != nil {
				return fmt.Errorf(
					"osapi client %s %s is both a hand-written command and the %s group's: "+
						"delete the hand-written one from cmd/",
					strings.Join(words, " "), c.Name(), name)
			}
			parent.AddCommand(c)
		}
	}

	return nil
}

func child(
	parent *cobra.Command,
	name string,
) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}

	return nil
}

// Render sets how a method's answer prints without --json. fn gets the
// command, for a flag that Flag added, the input the command sent and the
// output the API answered.
func Render[In, Out any](
	g Group,
	method string,
	fn func(cmd *cobra.Command, in In, out Out),
) {
	renderers[g.GroupName()+"."+method] = renderer{
		fn:  func(cmd *cobra.Command, in, out any) { fn(cmd, in.(In), out.(Out)) },
		out: reflect.TypeFor[Out](),
	}
}

// Acts sets what a command does with a method's answer, with or without
// --json: such as export, which writes the answer to a file. fn prints its
// own output, and reads --json from the command when it prints differently.
func Acts[In, Out any](
	g Group,
	method string,
	fn func(cmd *cobra.Command, in In, out Out),
) {
	Render(g, method, fn)
	r := renderers[g.GroupName()+"."+method]
	r.acts = true
	renderers[g.GroupName()+"."+method] = r
}

// Output prints the answer of method of group to the command that sent in:
// the body as it came for --json, or when the method has no renderer; else
// the answer decodes into the type of the method's renderer, which prints
// it. A renderer that Acts set runs with --json too.
func Output(
	cmd *cobra.Command,
	group string,
	method string,
	in any,
	answer []byte,
) {
	render, ok := renderers[group+"."+method]
	if deps.JSON() && !render.acts {
		deps.PrintJSON(answer)

		return
	}
	if !ok {
		deps.PrintJSON(answer)

		return
	}
	out := reflect.New(render.out)
	if len(answer) > 0 {
		if err := json.Unmarshal(answer, out.Interface()); err != nil {
			deps.Error(fmt.Errorf("decode %s %s: %w", group, method, err))

			return
		}
	}
	render.fn(cmd, in, out.Elem().Interface())
}

// Report reports a failed call of the running command, through the Error
// that Attach got.
func Report(
	err error,
) {
	deps.Error(err)
}

// ReadFile is the content of the local file at path, for a flag of a []byte
// field. A file that it cannot read is reported, and ok is false.
func ReadFile(
	path string,
) ([]byte, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		deps.Error(fmt.Errorf("read %s: %w", path, err))

		return nil, false
	}

	return data, true
}

// traceID is the trace of the call the running command made last. A
// command runs one call, so a renderer reads it with TraceID.
var traceID string

// SetTraceID keeps the trace of the call the running command made, for
// TraceID. The layer that sends the call sets it before Output.
func SetTraceID(
	id string,
) {
	traceID = id
}

// TraceID is the trace of the call that the running command made, from the
// controller's traceresponse header, for a renderer that prints it, as each
// behavior's result table does. It is "" when the answer had none.
func TraceID() string {
	return traceID
}
