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

package command_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/osapi-io/osapi/pkg/sdk/client"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/reference"
)

// group is a group of commands with no endpoint under it.
type group struct {
	name, doc, notes, long string
	path                   []string
	subs                   map[string]string
}

func (g group) GroupName() string                                   { return g.name }
func (g group) GroupDoc() string                                    { return g.doc }
func (g group) GroupNotes() string                                  { return g.notes }
func (g group) GroupLong() string                                   { return g.long }
func (g group) SubDoc(segment string) string                        { return g.subs[segment] }
func (g group) Commands() []command.Command                         { return nil }
func (g group) CommandPermission(command.Command) client.Permission { return "" }

func (g group) CommandPath() []string {
	if g.path != nil {
		return g.path
	}

	return []string{g.name}
}

// areaGroup is a group whose operations are in an area, as an endpoint
// group is.
type areaGroup struct {
	group
	area area.Name
}

func (g areaGroup) GroupArea() area.Name { return g.area }

type input struct {
	Name string
}

type CommandPublicTestSuite struct {
	suite.Suite
}

func (s *CommandPublicTestSuite) TestReadFile() {
	var reported error
	parent := &cobra.Command{Use: "client"}
	s.Require().NoError(command.Attach(parent, command.Deps{
		JSON:  func() bool { return false },
		Error: func(err error) { reported = err },
	}))

	path := filepath.Join(s.T().TempDir(), "a.conf")
	s.Require().NoError(os.WriteFile(path, []byte("data"), 0o600))
	data, ok := command.ReadFile(path)
	s.True(ok)
	s.Equal("data", string(data))

	_, ok = command.ReadFile(filepath.Join(s.T().TempDir(), "missing"))
	s.False(ok)
	s.ErrorContains(reported, "missing")
}

func (s *CommandPublicTestSuite) TestReport() {
	var reported error
	s.Require().NoError(command.Attach(&cobra.Command{Use: "client"}, command.Deps{
		Error: func(err error) { reported = err },
	}))

	command.Report(errors.New("api error (404): not found"))

	s.EqualError(reported, "api error (404): not found")
}

func (s *CommandPublicTestSuite) TestTraceID() {
	command.SetTraceID("4bf92f3577b34da6a3ce929d0e0e4736")
	s.Equal("4bf92f3577b34da6a3ce929d0e0e4736", command.TraceID())
	command.SetTraceID("")
	s.Empty(command.TraceID())
}

func (s *CommandPublicTestSuite) TestOutput() {
	g := group{name: "zzoutput"}

	tests := []struct {
		name         string
		method       string
		json         bool
		answer       []byte
		validateFunc func(printed string, err error)
	}{
		{
			name:   "when --json is given, the body prints as it came",
			method: "get",
			json:   true,
			answer: []byte(`{"name":"web-01"}`),
			validateFunc: func(printed string, err error) {
				s.Require().NoError(err)
				s.Equal(`{"name":"web-01"}`, printed)
			},
		},
		{
			name:   "when the method is not registered, the body prints as JSON",
			method: "set",
			answer: []byte(`{"name":"web-02"}`),
			validateFunc: func(printed string, err error) {
				s.Require().NoError(err)
				s.Equal(`{"name":"web-02"}`, printed)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var gotErr error
			printed := ""
			s.Require().NoError(command.Attach(&cobra.Command{Use: "client"}, command.Deps{
				JSON:      func() bool { return tc.json },
				Error:     func(err error) { gotErr = err },
				PrintJSON: func(body []byte) { printed = string(body) },
			}))

			command.Output(g, tc.method, input{Name: "web"}, tc.answer)

			tc.validateFunc(printed, gotErr)
		})
	}
}

// TestAttachUnderPrefix puts a group under the commands of its path, made
// when the parent has none and found when it has, with its notes.
// permGroup is a group whose commands each need the permission of their
// method name.
type permGroup struct {
	group
	cmds  []command.Command
	perms map[string]string
}

func (g permGroup) Commands() []command.Command { return g.cmds }

func (g permGroup) CommandPermission(m command.Command) client.Permission {
	return client.Permission(g.perms[m.MethodName()])
}

// TestAttachHelp gives the commands that Attach makes the help of their
// group: the group's Long, the one line of a Sub, and the permission that
// covers each command under it.
func (s *CommandPublicTestSuite) TestAttachHelp() {
	sub := func(name string) *cobra.Command {
		return &cobra.Command{Use: name, Annotations: map[string]string{command.SubKey: "trace"}}
	}
	same := permGroup{
		group: group{
			name: "zzhelp",
			doc:  "Help",
			long: "More.",
			subs: map[string]string{"trace": "Read a trace"},
		},
		cmds: []command.Command{
			command.Method[struct{}, struct{}]("get").Sub("trace"),
			command.Method[struct{}, struct{}]("list").Sub("trace"),
			command.Method[struct{}, struct{}]("info"),
		},
		perms: map[string]string{
			"trace get":  "trace:read",
			"trace list": "trace:read",
			"info":       "trace:read",
		},
	}
	command.AddCommands(same, sub("get"), sub("list"))
	mixed := permGroup{
		group: group{name: "zzmixed", doc: "Mixed"},
		cmds: []command.Command{
			command.Method[struct{}, struct{}]("get"),
			command.Method[struct{}, struct{}]("set"),
			command.Method[struct{}, struct{}]("ping"),
		},
		perms: map[string]string{"get": "x:read", "set": "x:write"},
	}
	command.AddCommands(mixed, &cobra.Command{Use: "get"})
	open := permGroup{
		group: group{name: "zzopen", doc: "Open"},
		cmds:  []command.Command{command.Method[struct{}, struct{}]("ping")},
	}
	command.AddCommands(open, &cobra.Command{Use: "ping"})

	parent := &cobra.Command{Use: "client"}
	s.Require().NoError(command.Attach(parent, command.Deps{}))

	found, _, err := parent.Find([]string{"zzhelp", "trace", "get"})
	s.Require().NoError(err)
	trace := found.Parent()
	s.Equal("Read a trace", trace.Short)
	s.Equal("trace:read", trace.Annotations[reference.Permission])
	s.Equal("More.", trace.Parent().Long)
	s.Equal("trace:read", trace.Parent().Annotations[reference.Permission],
		"one permission covers each command of the group")

	found, _, err = parent.Find([]string{"zzmixed", "get"})
	s.Require().NoError(err)
	s.Empty(found.Parent().Annotations[reference.Permission], "the commands need different ones")

	found, _, err = parent.Find([]string{"zzopen", "ping"})
	s.Require().NoError(err)
	s.Empty(found.Parent().Annotations[reference.Permission], "the command needs none")
}

func (s *CommandPublicTestSuite) TestAttachUnderPrefix() {
	g := group{
		name:  "zzbehavior",
		doc:   "A behavior",
		notes: "See its page.",
		path:  []string{"zznext", "zznode", "zzbehavior"},
	}
	get := &cobra.Command{Use: "get"}
	command.AddCommands(g, get)

	parent := &cobra.Command{Use: "client"}
	node := &cobra.Command{Use: "zznext"}
	parent.AddCommand(node)
	s.Require().NoError(command.Attach(parent, command.Deps{}))

	found, _, err := parent.Find([]string{"zznext", "zznode", "zzbehavior", "get"})
	s.Require().NoError(err)
	s.Same(get, found)
	s.Equal("Zznode commands", found.Parent().Parent().Short, "a missing segment is made")
	s.Same(node, found.Parent().Parent().Parent(), "a segment the parent has is used")
	s.Equal("See its page.", found.Parent().Annotations[reference.Notes])
}

// TestGroupByArea: the help of a command lists each group under the label of
// its area, in the order of area.All, and a parent whose child came first
// still gets its own help and area.
func (s *CommandPublicTestSuite) TestGroupByArea() {
	child := areaGroup{
		group: group{
			name: "zzpassword",
			doc:  "The password",
			path: []string{"zzarea", "zzuser", "zzpassword"},
		},
		area: area.Security,
	}
	command.AddCommands(child, &cobra.Command{Use: "update"})
	user := areaGroup{
		group: group{name: "zzuser", doc: "The users", path: []string{"zzarea", "zzuser"}},
		area:  area.Security,
	}
	command.AddCommands(user, &cobra.Command{Use: "list"})
	svc := areaGroup{
		group: group{name: "zzservice", doc: "The services", path: []string{"zzarea", "zzservice"}},
		area:  area.Services,
	}
	command.AddCommands(svc, &cobra.Command{Use: "list"})
	plain := group{name: "zzplain", doc: "No area", path: []string{"zzarea", "zzplain"}}
	command.AddCommands(plain, &cobra.Command{Use: "list"})

	parent := &cobra.Command{Use: "client"}
	s.Require().NoError(command.Attach(parent, command.Deps{}))
	node, _, err := parent.Find([]string{"zzarea"})
	s.Require().NoError(err)
	command.GroupByArea(node)

	found, _, err := parent.Find([]string{"zzarea", "zzuser"})
	s.Require().NoError(err)
	s.Equal("The users", found.Short, "the parent's group fills in the command its child made")
	s.Equal(string(area.Security), found.GroupID)

	found, _, err = parent.Find([]string{"zzarea", "zzplain"})
	s.Require().NoError(err)
	s.Empty(found.GroupID, "a group of no area stays under Additional Commands")

	titles := make([]string, 0, len(node.Groups()))
	for _, g := range node.Groups() {
		titles = append(titles, g.Title)
	}
	s.Equal([]string{"Services:", "Security:"}, titles, "the areas in the order of area.All")
}

func (s *CommandPublicTestSuite) TestAttach() {
	g := group{name: "zzattach", doc: "Attached things"}
	get := &cobra.Command{Use: "get"}
	command.AddCommands(g, get)

	parent := &cobra.Command{Use: "client"}
	s.Require().NoError(command.Attach(parent, command.Deps{}))
	var made *cobra.Command
	for _, c := range parent.Commands() {
		if c.Name() == "zzattach" {
			made = c
		}
	}
	s.Require().NotNil(made, "the group's command is made")
	s.Equal("Attached things", made.Short)
	s.Nil(made.Annotations, "a group with no notes has none")
	s.Contains(made.Commands(), get)

	// A parent that has the group, and a command of the same name, is a
	// command written twice.
	parent = &cobra.Command{Use: "client"}
	existing := &cobra.Command{Use: "zzattach"}
	existing.AddCommand(&cobra.Command{Use: "get"})
	parent.AddCommand(existing)
	err := command.Attach(parent, command.Deps{})
	s.Require().Error(err)
	s.ErrorContains(err, "osapi client zzattach get is both a hand-written command "+
		"and the zzattach group's")

	// A command with a Sub annotation goes under that command: a new one
	// when the group has none, the one there when it has.
	g2 := group{name: "zznested", doc: "Nested things"}
	nested := &cobra.Command{Use: "get", Annotations: map[string]string{command.SubKey: "trace"}}
	nested2 := &cobra.Command{Use: "list", Annotations: map[string]string{command.SubKey: "trace"}}
	command.AddCommands(g2, nested, nested2)
	parent = &cobra.Command{Use: "client"}
	s.Require().NoError(command.Attach(parent, command.Deps{}))
	var trace *cobra.Command
	for _, c := range parent.Commands() {
		if c.Name() == "zznested" {
			for _, sub := range c.Commands() {
				if sub.Name() == "trace" {
					trace = sub
				}
			}
		}
	}
	s.Require().NotNil(trace, "the sub command is made")
	s.Equal("Trace commands", trace.Short)
	s.Len(trace.Commands(), 2)

	// A named sub command that is not under its group yet gets them.
	parent = &cobra.Command{Use: "client"}
	nestedGroup := &cobra.Command{Use: "zznested"}
	namedTrace := &cobra.Command{Use: "trace", Short: "Read a trace"}
	s.Require().NoError(command.Attach(parent, command.Deps{}, nestedGroup, namedTrace))
	s.Len(namedTrace.Commands(), 2)
	for _, c := range nestedGroup.Commands() {
		s.NotEqual("trace", c.Name(), "the named trace gets them, not a second one")
	}

	// A hand-written group command that is named, but not under parent yet,
	// gets the commands, and parent gets no second one.
	parent = &cobra.Command{Use: "client"}
	named := &cobra.Command{Use: "zzattach"}
	s.Require().NoError(command.Attach(parent, command.Deps{}, named))
	for _, c := range parent.Commands() {
		s.NotEqual("zzattach", c.Name(), "the named command gets them, not a second one")
	}
	s.Len(named.Commands(), 1)
}

// TestAttachLocalFlat puts each command of a flat group under the group's
// parent, the root or the words of Under, with no word for the group.
func (s *CommandPublicTestSuite) TestAttachLocalFlat() {
	atRoot := command.Define("zzflatroot", "At the root", command.Flat())
	run := &cobra.Command{Use: "zzrun"}
	command.AddCommands(atRoot, run)
	under := command.Define("zzflatunder", "Under", command.Flat(), command.Under("zzflatparent"))
	start := &cobra.Command{Use: "start"}
	command.AddCommands(under, start)

	s.Empty(atRoot.CommandPath())
	s.Equal("zzflatroot", atRoot.Path())
	s.Equal([]string{"zzflatparent"}, under.CommandPath())
	s.Equal("zzflatparent zzflatunder", under.Path())

	root := &cobra.Command{Use: "osapi"}
	s.Require().NoError(command.AttachLocal(root, command.Deps{}))
	found, _, err := root.Find([]string{"zzrun"})
	s.Require().NoError(err)
	s.Same(run, found, "a flat group's command is under the root")
	origin, ok := command.Origin(run)
	s.True(ok)
	s.Same(atRoot, origin)
	found, _, err = root.Find([]string{"zzflatparent", "start"})
	s.Require().NoError(err)
	s.Same(start, found, "a flat group's command is under the words of Under")

	// A command that is there already is a command made twice.
	again := &cobra.Command{Use: "osapi"}
	again.AddCommand(&cobra.Command{Use: "zzrun"})
	s.ErrorContains(command.AttachLocal(again, command.Deps{}), "osapi zzrun is made twice")
}

// TestAttachLocal puts a local group under the root, under the words of
// Under, and leaves it out of Attach; Origin names the group of each
// command it added.
func (s *CommandPublicTestSuite) TestAttachLocal() {
	g := command.Define("zzlocalattach", "Local things", command.Under("zzup"))
	get := &cobra.Command{Use: "get"}
	command.AddCommands(g, get)

	client := &cobra.Command{Use: "client"}
	s.Require().NoError(command.Attach(client, command.Deps{}))
	_, ok := command.Origin(get)
	s.False(ok, "Attach leaves a local group out")

	// A named command for a word of Under is used, and is added by the code
	// that wrote it.
	root := &cobra.Command{Use: "osapi"}
	up := &cobra.Command{Use: "zzup", Short: "Up things"}
	s.Require().NoError(command.AttachLocal(root, command.Deps{}, up))
	s.Empty(root.Commands(), "the named parent is not added here")
	root.AddCommand(up)
	found, _, err := root.Find([]string{"zzup", "zzlocalattach", "get"})
	s.Require().NoError(err)
	s.Same(get, found)
	origin, ok := command.Origin(get)
	s.True(ok)
	s.Same(g, origin)
	_, ok = command.Origin(root)
	s.False(ok, "a hand-written command has no group")

	// A command of the same name under the root is a command written twice.
	root = &cobra.Command{Use: "osapi"}
	group := &cobra.Command{Use: "zzlocalattach"}
	group.AddCommand(&cobra.Command{Use: "get"})
	up = &cobra.Command{Use: "zzup"}
	up.AddCommand(group)
	root.AddCommand(up)
	err = command.AttachLocal(root, command.Deps{})
	s.ErrorContains(err, "osapi zzup zzlocalattach get is both a hand-written command "+
		"and the zzlocalattach group's")
}

func TestCommandPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CommandPublicTestSuite))
}
