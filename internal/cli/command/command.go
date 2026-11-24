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

// Package command is the command layer of the CLI, and the DSL of its local
// commands. A group's gen/cli.go makes one cobra command for each method and
// registers it here: Attach puts the commands of the endpoint groups and the
// behaviors under osapi client, and AttachLocal puts the commands of the
// local groups under osapi. A local group, which Define makes, runs in the
// CLI and calls no API: Run checks its input and calls its handler. The
// layer that sends an endpoint command's request gives the answer to Output,
// as Run gives a local method's output, and Output prints it: as JSON for
// --json, or from the method's output type, by the output style of the CLI.
// No command writes its own output.
package command

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/cli/reference"
)

// Group is a group of commands, osapi client <group>, with one command for
// each of its methods. An endpoint group is one.
type Group interface {
	// GroupName is the group's command.
	GroupName() string
	// GroupDoc says what the group is, in one line: its command's help.
	GroupDoc() string
	// GroupNotes is the Notes section of the page of the group's command,
	// or "".
	GroupNotes() string
	// GroupLong is the rest of the help of the group's command, or "".
	GroupLong() string
	// SubDoc is the one line of help of the command that a Sub segment
	// makes, such as trace, or "".
	SubDoc(segment string) string
	// CommandPath is the words of the group's command under osapi client,
	// such as next node sysctl.
	CommandPath() []string
	// Commands are the group's methods, in the order they were declared.
	Commands() []Command
	// CommandPermission is the permission the command of method m needs, or
	// "" for one that needs none.
	CommandPermission(m Command) types.Permission
}

// Command is one method of a group: one command.
type Command interface {
	// MethodName is the method's name in its group.
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
	// PageNotes is the Notes section of the command's page, or "".
	PageNotes() string
	// PageResult is the Result section of the command's page, or "".
	PageResult() string
	// Output is the type of the method's output, which the output prints
	// from.
	Output() reflect.Type
	// KeyField is the json name of the input field that names what the
	// command acts on, such as machine_id, or "".
	KeyField() string
	// Saying is what the one line of a method with no body says, which Says
	// set, or "" for the past tense of the method.
	Saying() string
}

// Example is an example of a CLI command: its flags, and what it shows,
// which the help writes as a # line above it.
type Example struct {
	Flags string
	Doc   string
	// Output is what the API answers for the example, a value of the
	// method's output type, which Prints sets. The docs page of the command
	// shows it as the CLI prints it. Nil shows no output.
	Output any
	// Logs are the lines that a process writes when it starts, which Logs
	// sets. The docs page of the command shows them as its log handler
	// writes them.
	Logs []LogLine
}

// LogLine is a line that a process writes to its log: its level, its
// message, and its attributes in order, such as component=nats.
type LogLine struct {
	Level slog.Level
	Msg   string
	Attrs []slog.Attr
}

// Deps is what the commands need from the CLI.
type Deps struct {
	// JSON reports whether --json was given.
	JSON func() bool
	// Error reports a failed call.
	Error func(err error)
	// PrintJSON prints an answer as it came, for --json.
	PrintJSON func(body []byte)
	// Print is what the output of a method prints through: the CLI's
	// headings, blocks, tables and formats.
	Print Printer
}

type entry struct {
	group Group
	cmds  []*cobra.Command
}

var (
	registered []entry
	deps       Deps
	// printers print the answers of a kind of group, such as the result
	// table of each behavior. PrintsGroups adds one.
	printers []GroupPrinter
)

// SubKey is the annotation of a generated command that goes under a
// command of its own: the Sub segment of its method, such as trace in
// osapi client trace get.
const SubKey = "osapi.endpoint.sub"

// AddCommands adds a group's generated commands. Each gen/cli.go calls it
// from init.
func AddCommands(
	g Group,
	cmds ...*cobra.Command,
) {
	registered = append(registered, entry{group: g, cmds: cmds})
}

// annotate sets the annotation key of cmd to value, when value is not "".
func annotate(
	cmd *cobra.Command,
	key string,
	value string,
) {
	if value == "" {
		return
	}
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[key] = value
}

// permissionOf is the permission that each command of g under the Sub
// segment needs, or each command of g for "", when they all need the same
// one. It is "" when they need different ones, or none: the page of a
// parent names a permission only when one covers each command under it.
func permissionOf(
	g Group,
	sub string,
) types.Permission {
	var perm types.Permission
	for _, c := range g.Commands() {
		if sub != "" && c.SubPath() != sub {
			continue
		}
		p := g.CommandPermission(c)
		if p == "" || perm != "" && p != perm {
			return ""
		}
		perm = p
	}

	return perm
}

// Attach adds every registered group's commands under parent, osapi client,
// and keeps d for the commands to call through. A group goes under the
// commands of its command path: next node sysctl is under next, then node,
// which Attach adds when parent does not have them. A group or a command
// that is among groups, by name, is that one, so the generated commands
// join a hand-written command of the same name, such as agent with key and
// status. A local group is not attached here: AttachLocal puts it under the
// root.
func Attach(
	parent *cobra.Command,
	d Deps,
	groups ...*cobra.Command,
) error {
	return attach(parent, d, false, groups)
}

// AttachLocal adds the commands of every registered local group under root,
// osapi, as Attach does for the groups under osapi client: a group goes
// under the commands that Under names, such as controller.
func AttachLocal(
	root *cobra.Command,
	d Deps,
	groups ...*cobra.Command,
) error {
	return attach(root, d, true, groups)
}

func attach(
	parent *cobra.Command,
	d Deps,
	local bool,
	groups []*cobra.Command,
) error {
	deps = d
	for _, e := range registered {
		if _, ok := e.group.(*LocalGroup); ok != local {
			continue
		}
		name := e.group.GroupName()
		words := e.group.CommandPath()
		// A flat group adds no word: its path is its parents, and each of
		// its commands is under them, such as osapi version under the root.
		lg, isLocal := e.group.(*LocalGroup)
		flat := isLocal && lg.Flat
		parents := words
		if !flat {
			parents = words[:len(words)-1]
		}
		into := parent
		for _, w := range parents {
			next := child(into, w)
			// A named command that is not under into yet is added by the
			// code that wrote it.
			for _, g := range groups {
				if next == nil && g.Name() == w {
					next = g
				}
			}
			if next == nil {
				next = &cobra.Command{Use: w, Short: strings.ToUpper(w[:1]) + w[1:] + " commands"}
				into.AddCommand(next)
				// A child's group can come before its parent's, as
				// password comes before user: the parent's group fills
				// this command in when it comes.
				placeholders[next] = true
			}
			into = next
		}
		if flat {
			for _, c := range e.cmds {
				if child(into, c.Name()) != nil {
					return fmt.Errorf(
						"%s %s is made twice, by hand or by two groups: keep one",
						prefix(local), strings.Join(append(parents, c.Name()), " "))
				}
				into.AddCommand(c)
				origins[c] = e.group
			}

			continue
		}
		group := child(into, name)
		for _, g := range groups {
			if g.Name() == name {
				group = g
			}
		}
		if group == nil {
			group = &cobra.Command{Use: name}
			into.AddCommand(group)
			placeholders[group] = true
		}
		if placeholders[group] {
			delete(placeholders, group)
			group.Short = e.group.GroupDoc()
			group.Long = e.group.GroupLong()
			annotate(group, reference.Notes, e.group.GroupNotes())
			if ag, ok := e.group.(areaGroup); ok && ag.GroupArea() != "" {
				areas[group] = ag.GroupArea()
			}
			annotate(group, reference.Permission, string(permissionOf(e.group, "")))
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
					short := e.group.SubDoc(sub)
					if short == "" {
						short = strings.ToUpper(sub[:1]) + sub[1:] + " commands"
					}
					parent = &cobra.Command{Use: sub, Short: short}
					annotate(parent, reference.Permission, string(permissionOf(e.group, sub)))
					group.AddCommand(parent)
				}
			}
			if child(parent, c.Name()) != nil {
				return fmt.Errorf(
					"%s %s %s is both a hand-written command and the %s group's: "+
						"delete the hand-written one from cmd/",
					prefix(local), strings.Join(words, " "), c.Name(), name)
			}
			parent.AddCommand(c)
			origins[c] = e.group
		}
	}

	return nil
}

// prefix is the command that a group's command path is under: osapi for a
// local group, and osapi client for the others.
func prefix(
	local bool,
) string {
	if local {
		return "osapi"
	}

	return "osapi client"
}

// origins is the group that made each attached command.
var origins = map[*cobra.Command]Group{}

// areaGroup is a group whose operations are in an area, one of area.All,
// as each endpoint group is.
type areaGroup interface {
	GroupArea() area.Name
}

// areas is the area of each group's command that Attach made.
var areas = map[*cobra.Command]area.Name{}

// placeholders are the commands that Attach made for a word of a command
// path whose group it has not attached yet.
var placeholders = map[*cobra.Command]bool{}

// GroupByArea puts each command under cmd that Attach made for a group in a
// cobra command group of the group's area, so the help of cmd lists them
// under the label of each area, in the order of area.All: osapi client node
// --help lists dns and interface under Networking. A command of no area
// stays under Additional Commands. Call it after Attach.
func GroupByArea(
	cmd *cobra.Command,
) {
	used := map[area.Name]bool{}
	for _, c := range cmd.Commands() {
		if a, ok := areas[c]; ok {
			c.GroupID = string(a)
			used[a] = true
		}
	}
	for _, a := range area.All {
		if used[a.Name] && !cmd.ContainsGroup(string(a.Name)) {
			cmd.AddGroup(&cobra.Group{ID: string(a.Name), Title: a.Label + ":"})
		}
	}
}

// Origin is the group whose definition made c, a command that Attach or
// AttachLocal added, and false for a command that a person wrote.
func Origin(
	c *cobra.Command,
) (Group, bool) {
	g, ok := origins[c]

	return g, ok
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

// GroupPrinter prints the answer of method of group g, and is true, or is
// false for a group that it does not print.
type GroupPrinter func(g Group, method string, answer []byte) bool

// PrintsGroups adds p, which prints a kind of group. The behavior runtime
// adds the result table of each behavior.
func PrintsGroups(
	p GroupPrinter,
) {
	printers = append(printers, p)
}

// Groups are the registered groups: each endpoint group, behavior and
// local group whose commands AddCommands added, in the order they were
// added.
func Groups() []Group {
	out := make([]Group, 0, len(registered))
	for _, e := range registered {
		out = append(out, e.group)
	}

	return out
}

// lookup is the command of the method of the registered group g, and its
// group, or nil. Two groups can have one name, such as the controller's
// file endpoints and the file behavior under node, so a group that is a
// pointer is found by its identity, and another by its name.
func lookup(
	group Group,
	name string,
) (Group, Command) {
	for _, g := range Groups() {
		if !same(g, group) {
			continue
		}
		for _, c := range g.Commands() {
			if c.MethodName() == name {
				return g, c
			}
		}
	}

	return nil, nil
}

// same reports whether a and b are one group: the same pointer, or, for a
// group that is not a pointer, the same name.
func same(
	a Group,
	b Group,
) bool {
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	if va.Kind() == reflect.Pointer && vb.Kind() == reflect.Pointer {
		return va.Pointer() == vb.Pointer()
	}

	return a.GroupName() == b.GroupName()
}

// Output prints the answer of method of group, to the input in. With --json
// it prints the body as it came, or, for a method that answered with no
// body, its key and what it did. Without it, a group that a GroupPrinter
// prints, such as a behavior, prints its way, and each other method prints
// from its output type.
func Output(
	group Group,
	method string,
	in any,
	answer []byte,
) {
	if deps.JSON() {
		if len(answer) > 0 {
			deps.PrintJSON(answer)

			return
		}
		// A method that answers with no body still says what occurred, so
		// a script reads it as it reads any answer.
		if g, m := lookup(group, method); g != nil {
			done, _ := json.Marshal(
				map[string]string{keyName(m): keyOf(g, m, in), "result": said(m)},
			)
			deps.PrintJSON(done)
		}

		return
	}
	for _, p := range printers {
		if p(group, method, answer) {
			return
		}
	}
	derived(group, method, in, answer)
}

// derived prints the answer from the output type of the method, or as it
// came for a method that is not registered.
func derived(
	group Group,
	name string,
	in any,
	answer []byte,
) {
	g, m := lookup(group, name)
	if m == nil {
		deps.PrintJSON(answer)

		return
	}
	if err := derive(deps.Print, g, m, in, answer); err != nil {
		deps.Error(err)
	}
}

// Report reports a failed call of the running command, through the Error
// that Attach got.
func Report(
	err error,
) {
	deps.Error(err)
}

// Strings is values as a list of T, for a flag of a []T field whose T is a
// named string type, such as []types.Role.
func Strings[T ~string](
	values []string,
) []T {
	out := make([]T, 0, len(values))
	for _, v := range values {
		out = append(out, T(v))
	}

	return out
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
// command runs one call, so its printer reads it with TraceID.
var traceID string

// SetTraceID keeps the trace of the call the running command made, for
// TraceID. The layer that sends the call sets it before Output.
func SetTraceID(
	id string,
) {
	traceID = id
}

// TraceID is the trace of the call that the running command made, from the
// controller's traceresponse header, for a printer that prints it, as each
// behavior's result table does. It is "" when the answer had none.
func TraceID() string {
	return traceID
}
