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

package commanddsl_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/osapi-io/osapi/pkg/sdk/client"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/commanddsl"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

type widgetIn struct {
	_    struct{}
	Name string `json:"name" validate:"required" doc:"The widget" short:"n"`
}

type widgetOut struct {
	Name string `json:"name"`
}

func widgetGet(context.Context, struct{}, widgetIn) (widgetOut, error) {
	return widgetOut{}, nil
}

func widgetList(context.Context, struct{}, struct{}) ([]widgetOut, error) {
	return nil, nil
}

// apiGroup is a group of the endpoint or the behavior layer: its commands
// call the API.
type apiGroup struct {
	name string
}

func (g apiGroup) GroupName() string                                   { return g.name }
func (g apiGroup) GroupDoc() string                                    { return "Remote " + g.name }
func (g apiGroup) GroupNotes() string                                  { return "" }
func (g apiGroup) GroupLong() string                                   { return "" }
func (g apiGroup) SubDoc(string) string                                { return "" }
func (g apiGroup) CommandPath() []string                               { return []string{g.name} }
func (g apiGroup) CommandPermission(command.Command) client.Permission { return "" }
func (g apiGroup) Commands() []command.Command {
	return []command.Command{command.Method[struct{}, widgetOut]("get").Short("Get it")}
}

var (
	widget = command.Register(command.Define("zzwidget", "Widgets",
		command.Uses[struct{}](),
		command.Method[widgetIn, widgetOut]("get").Short("Get a widget").
			Example("--name a").Prints(widgetOut{}).Fails("no name", widgetIn{}).Handle(widgetGet),
		command.Method[struct{}, []widgetOut]("list").Short("List widgets").
			Example("").Prints([]widgetOut{}).Handle(widgetList),
	))
	root = &cobra.Command{Use: "osapi"}

	// zzmore has a process, a method that says a state, and one under a Sub.
	_ = command.Register(command.Define("zzmore", "More",
		command.Uses[struct{}](),
		command.SubDoc("keys", "The keys"),
		command.Process("start").Short("Start it").Flag("controller.api.port").
			Setup(func(context.Context, struct{}) (command.Running, error) { return command.Running{}, nil }),
		command.Method[widgetIn, struct{}]("drain").Short("Drain it").Says("draining").
			Example("--name a").Fails("no name", widgetIn{}).
			Handle(func(context.Context, struct{}, widgetIn) (struct{}, error) { return struct{}{}, nil }),
		command.Method[struct{}, widgetOut]("get").Sub("keys").Short("Get a key").Example("").
			Prints(widgetOut{}).
			Handle(func(context.Context, struct{}, struct{}) (widgetOut, error) { return widgetOut{}, nil }),
	))
)

func (s *MethodsPublicTestSuite) SetupSuite() {
	command.AddCommands(widget, &cobra.Command{Use: "get", Run: func(*cobra.Command, []string) {}},
		&cobra.Command{Use: "list", Run: func(*cobra.Command, []string) {}})
	s.Require().NoError(command.AttachLocal(root, command.Deps{}))
	client := &cobra.Command{Use: "client"}
	root.AddCommand(client)
	for _, name := range []string{"zzremote", "zzbehave"} {
		command.AddCommands(
			apiGroup{name: name},
			&cobra.Command{Use: "get", Run: func(*cobra.Command, []string) {}},
		)
	}
	s.Require().NoError(command.Attach(client, command.Deps{}))
	root.AddCommand(
		&cobra.Command{Use: "hand", Run: func(*cobra.Command, []string) {}},
		&cobra.Command{Use: "hidden", Hidden: true, Run: func(*cobra.Command, []string) {}},
		&cobra.Command{Use: "help", Run: func(*cobra.Command, []string) {}},
	)
}

func (s *MethodsPublicTestSuite) deps() commanddsl.Deps {
	return commanddsl.Deps{
		Root:       root,
		IsBehavior: func(name string) bool { return name == "zzbehave" },
	}
}

func (s *MethodsPublicTestSuite) TestMethodGet() {
	tests := []struct {
		name         string
		in           commanddsl.MethodInput
		validateFunc func(commanddsl.MethodOutput, error)
	}{
		{
			name: "when the method has flags, each has its check, with the output and examples",
			in:   commanddsl.MethodInput{Name: "zzwidget", Method: "get"},
			validateFunc: func(got commanddsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Equal("osapi zzwidget get", got.Command)
				s.Equal("Get a widget", got.Doc)
				s.Equal("--name", got.Flags[0].Flag)
				s.Equal("required", got.Flags[0].Validate)
				s.Equal("widgetOut", got.OutputType)
				s.Equal("name", got.Results[0].Field)
				s.Equal([]commanddsl.Example{{Flags: "--name a"}}, got.Examples)
				s.Equal([]string{"no name"}, got.Refuses)
				s.NotNil(got.Output)
			},
		},
		{
			name: "when the method is a process, it has its config keys and no output",
			in:   commanddsl.MethodInput{Name: "zzmore", Method: "start"},
			validateFunc: func(got commanddsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.True(got.Process)
				s.Equal([]string{"controller.api.port"}, got.ConfigKeys)
				s.Empty(got.OutputType)
				s.Equal([]commanddsl.Flag{}, got.Flags)
			},
		},
		{
			name: "when the method says a state, and is under a Sub, both show",
			in:   commanddsl.MethodInput{Name: "zzmore", Method: "keys get"},
			validateFunc: func(got commanddsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Equal("osapi zzmore keys get", got.Command)
			},
		},
		{
			name: "when the method says a state, it is its line",
			in:   commanddsl.MethodInput{Name: "zzmore", Method: "drain"},
			validateFunc: func(got commanddsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Equal("draining", got.Says)
			},
		},
		{
			name: "when the group calls the API, it points at its own method get",
			in:   commanddsl.MethodInput{Name: "zzremote", Method: "get"},
			validateFunc: func(_ commanddsl.MethodOutput, err error) {
				s.ErrorContains(err, "osapi controller endpoint method get")
			},
		},
		{
			name: "when the group has no such method, it says how to list them",
			in:   commanddsl.MethodInput{Name: "zzwidget", Method: "nope"},
			validateFunc: func(_ commanddsl.MethodOutput, err error) {
				s.ErrorContains(err, "osapi cli command get --name zzwidget")
			},
		},
		{
			name: "when there is no such group, it says how to list them",
			in:   commanddsl.MethodInput{Name: "nope", Method: "get"},
			validateFunc: func(_ commanddsl.MethodOutput, err error) {
				s.ErrorContains(err, "osapi cli command list")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(commanddsl.MethodGet(context.Background(), s.deps(), tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestList() {
	got, err := commanddsl.List(context.Background(), s.deps(), commanddsl.ListInput{})
	s.Require().NoError(err)

	layers := map[string]string{}
	for _, it := range got.Items {
		layers[it.Command] = it.Layer
	}
	s.Equal(commanddsl.LayerCommand, layers["osapi zzwidget get"])
	s.Equal(commanddsl.LayerEndpoint, layers["osapi client zzremote get"])
	s.Equal(commanddsl.LayerBehavior, layers["osapi client zzbehave get"])
	s.Equal(commanddsl.LayerHandWritten, layers["osapi hand"])
	s.NotContains(layers, "osapi hidden")
	s.NotContains(layers, "osapi help")
	s.Equal(2, got.Command)
	s.Equal(1, got.Endpoint)
	s.Equal(1, got.Behavior)
	s.Equal(1, got.HandWritten)
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		in           commanddsl.GetInput
		validateFunc func(commanddsl.GetOutput, error)
	}{
		{
			name: "when the group is local, each command has its flags and refusals",
			in:   commanddsl.GetInput{Name: "zzwidget"},
			validateFunc: func(got commanddsl.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("zzwidget", got.Name)
				s.Require().Len(got.Commands, 2)
				s.Equal("osapi zzwidget get", got.Commands[0].Command)
				s.Equal([]string{"no name"}, got.Commands[0].Fails)
				s.Equal([]commanddsl.Flag{
					{
						Flag:     "--name",
						Short:    "n",
						Type:     "string",
						Validate: "required",
						Doc:      "The widget",
					},
				}, got.Commands[0].Flags)
			},
		},
		{
			name: "when the group calls the API, its commands are under osapi client",
			in:   commanddsl.GetInput{Name: "zzremote"},
			validateFunc: func(got commanddsl.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("osapi client zzremote get", got.Commands[0].Command)
				s.Empty(got.Commands[0].Flags)
			},
		},
		{
			name: "when no group has the name, it says to list them",
			in:   commanddsl.GetInput{Name: "nosuch"},
			validateFunc: func(_ commanddsl.GetOutput, err error) {
				s.ErrorContains(err, "osapi cli command list")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(commanddsl.Get(context.Background(), s.deps(), tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestCreate() {
	dir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module github.com/osapi-io/osapi\n\ngo 1.25\n"), 0o644))
	all := filepath.Join(dir, "internal", "cli", "commands", "all")
	s.Require().NoError(os.MkdirAll(all, 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(all, "all.go"),
		[]byte("// Package all registers every local group.\npackage all\n\nimport (\n)\n"), 0o644))
	d := s.deps()
	d.Dir = dir

	got, err := commanddsl.Create(context.Background(), d,
		commanddsl.CreateInput{Name: "gizmo", Description: "Gizmos"})
	s.Require().NoError(err)
	s.Equal("gizmo", got.Name)
	s.NotEmpty(got.Files)
	s.Contains(got.Next, commanddsl.Step{Command: "go test ./internal/cli/commands/gizmo/..."})

	_, err = commanddsl.Create(context.Background(), d,
		commanddsl.CreateInput{Name: "Bad", Description: "x"})
	s.Error(err)
}

func (s *MethodsPublicTestSuite) TestSpec() {
	got, err := commanddsl.Spec(context.Background(), s.deps(), commanddsl.SpecInput{})
	s.Require().NoError(err)
	s.Equal(command.Spec.Sections(), got.Kinds)
}

// TestValidate runs last: a broken group, once registered, stays.
func (s *MethodsPublicTestSuite) TestValidate() {
	got, err := commanddsl.Validate(context.Background(), s.deps(), commanddsl.ValidateInput{})
	s.Require().NoError(err)
	s.True(got.Valid)
	s.Positive(got.Local)

	command.Register(command.Define("zzbroken", "Broken", command.Uses[struct{}](),
		command.Method[struct{}, widgetOut]("frob").Short("X").
			Handle(func(context.Context, struct{}, struct{}) (widgetOut, error) {
				return widgetOut{}, nil
			})))
	got, err = commanddsl.Validate(context.Background(), s.deps(), commanddsl.ValidateInput{})
	s.ErrorIs(err, command.FailedError{})

	// A method whose output has no schema breaks a rule, and method get
	// says so too.
	command.Register(command.Define("zzbadout", "Bad output", command.Uses[struct{}](),
		command.Method[struct{}, chan int]("list").Short("List them")))
	_, mErr := commanddsl.MethodGet(context.Background(), s.deps(),
		commanddsl.MethodInput{Name: "zzbadout", Method: "list"})
	s.ErrorContains(mErr, "zzbadout list output")
	s.False(got.Valid)
	s.NotEmpty(got.Problems)
	s.Contains(got.Fixes, command.RuleVerb)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
