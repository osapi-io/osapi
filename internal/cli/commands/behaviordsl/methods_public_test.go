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

package behaviordsl_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/behaviordsl"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/job"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

type okIn struct {
	Key string `json:"key" validate:"required" doc:"The key"`
}

// broken are behaviors whose types have no JSON Schema: bad in has a
// channel for its input, and bad out for its output.
var broken = []*behavior.Behavior{
	{Name: "badin", Methods: []*behavior.Method{
		{Name: "get", In: reflect.TypeFor[chan int](), Out: reflect.TypeFor[okIn]()},
	}},
	{Name: "badout", Methods: []*behavior.Method{
		{Name: "get", In: reflect.TypeFor[okIn](), Out: reflect.TypeFor[chan int]()},
	}},
}

func (s *MethodsPublicTestSuite) deps() behaviordsl.Deps {
	return behaviordsl.Deps{
		Behaviors: func() []*behavior.Behavior { return append(behavior.All(), broken...) },
		Validate:  behavior.Validate,
		Keys:      config.Keys,
		Spec:      func() guide.Spec { return behavior.CurrentSpec(job.FactKeys()...) },
		OpenAPI: func(b *behavior.Behavior) (any, error) {
			if b.Name == "hostname" {
				return nil, errors.New("no schema")
			}

			return map[string]any{"openapi": "3.0.0", "info": map[string]any{"title": b.Name}}, nil
		},
	}
}

func (s *MethodsPublicTestSuite) TestList() {
	got, err := behaviordsl.List(context.Background(), s.deps(), behaviordsl.ListInput{})
	s.Require().NoError(err)
	byName := map[string]behaviordsl.Summary{}
	for _, it := range got.Items {
		byName[it.Name] = it
	}
	s.Contains(byName["sysctl"].Methods, "update")
	s.Equal(area.Config, byName["sysctl"].Area)
	s.True(byName["managedfile"].Internal)
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		in           behaviordsl.GetInput
		validateFunc func(behaviordsl.GetOutput, error)
	}{
		{
			name: "when the behavior is served, each method has its route, permission and command",
			in:   behaviordsl.GetInput{Name: "sysctl"},
			validateFunc: func(got behaviordsl.GetOutput, err error) {
				s.Require().NoError(err)
				s.True(got.API)
				s.Equal(area.Config, got.Area)
				s.Equal([]string{"managedfile"}, got.Uses)
				s.NotEmpty(got.Methods[0].Route)
				s.NotEmpty(got.Methods[0].Command)
				s.Equal("sysctl:read", got.Methods[0].Permission)
				s.Contains(got.Types[2].Input, "key*", "a required input is marked")
				s.Equal("behaviors.sysctl.dir", got.Settings[0].Key)
				s.Contains(got.Telemetry.Recorded, "key (")
			},
		},
		{
			name: "when the providers run a program with sudo, the behavior lists it once",
			in:   behaviordsl.GetInput{Name: "certificate"},
			validateFunc: func(got behaviordsl.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal([]string{"update-ca-certificates", "update-ca-trust"}, got.Sudo)
			},
		},
		{
			name: "when the behavior is internal, its methods have no route or command",
			in:   behaviordsl.GetInput{Name: "managedfile"},
			validateFunc: func(got behaviordsl.GetOutput, err error) {
				s.Require().NoError(err)
				s.True(got.Internal)
				s.Contains(got.UsedBy, "sysctl")
				s.Empty(got.Methods[0].Route)
				s.Empty(got.Methods[0].Command)
			},
		},
		{
			name: "when no behavior has the name, it says to list them",
			in:   behaviordsl.GetInput{Name: "nosuch"},
			validateFunc: func(_ behaviordsl.GetOutput, err error) {
				s.ErrorContains(err, "osapi dsl behavior list")
			},
		},
		{
			name: "when an input has no schema, it is the error",
			in:   behaviordsl.GetInput{Name: "badin"},
			validateFunc: func(_ behaviordsl.GetOutput, err error) {
				s.ErrorContains(err, "badin get input")
			},
		},
		{
			name: "when an output has no schema, it is the error",
			in:   behaviordsl.GetInput{Name: "badout"},
			validateFunc: func(_ behaviordsl.GetOutput, err error) {
				s.ErrorContains(err, "badout get output")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(behaviordsl.Get(context.Background(), s.deps(), tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestMethodGet() {
	tests := []struct {
		name         string
		in           behaviordsl.MethodInput
		validateFunc func(behaviordsl.MethodOutput, error)
	}{
		{
			name: "when the method is served, it has its fields, schemas and examples",
			in:   behaviordsl.MethodInput{Name: "sysctl", Method: "update"},
			validateFunc: func(got behaviordsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Equal("PUT", got.HTTPMethod)
				s.NotEmpty(got.Command)
				s.Equal("key", got.Fields[0].Field)
				s.True(got.Fields[0].Required)
				s.NotEmpty(got.Fields[0].Constraints)
				s.Equal("changed", got.Results[1].Field)
				s.NotNil(got.Input)
				s.NotEmpty(got.Examples)
			},
		},
		{
			name: "when the method lists, its output is described by its element",
			in:   behaviordsl.MethodInput{Name: "sysctl", Method: "list"},
			validateFunc: func(got behaviordsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Equal("[]Entry", got.OutputType)
				s.Equal("key", got.Results[0].Field)
			},
		},
		{
			name: "when the behavior is internal, the method is not served",
			in:   behaviordsl.MethodInput{Name: "managedfile", Method: "get"},
			validateFunc: func(got behaviordsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Empty(got.Route)
				s.Empty(got.Command)
			},
		},
		{
			name: "when the behavior has no such method, it says how to list them",
			in:   behaviordsl.MethodInput{Name: "sysctl", Method: "nosuch"},
			validateFunc: func(_ behaviordsl.MethodOutput, err error) {
				s.ErrorContains(err, "osapi dsl behavior get --name sysctl")
			},
		},
		{
			name: "when no behavior has the name, it is the error",
			in:   behaviordsl.MethodInput{Name: "nosuch", Method: "get"},
			validateFunc: func(_ behaviordsl.MethodOutput, err error) {
				s.ErrorContains(err, "no behavior named")
			},
		},
		{
			name: "when the input has no schema, it is the error",
			in:   behaviordsl.MethodInput{Name: "badin", Method: "get"},
			validateFunc: func(_ behaviordsl.MethodOutput, err error) {
				s.ErrorContains(err, "badin get input")
			},
		},
		{
			name: "when the output has no schema, it is the error",
			in:   behaviordsl.MethodInput{Name: "badout", Method: "get"},
			validateFunc: func(_ behaviordsl.MethodOutput, err error) {
				s.ErrorContains(err, "badout get output")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(behaviordsl.MethodGet(context.Background(), s.deps(), tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestOpenAPIGet() {
	tests := []struct {
		name         string
		in           behaviordsl.GetInput
		validateFunc func(behaviordsl.OpenAPIOutput, error)
	}{
		{
			name: "when the behavior is served, it is its part of the spec as YAML",
			in:   behaviordsl.GetInput{Name: "sysctl"},
			validateFunc: func(got behaviordsl.OpenAPIOutput, err error) {
				s.Require().NoError(err)
				s.Equal("sysctl", got.Name)
				s.Contains(got.YAML, "openapi: 3.0.0")
				s.Contains(got.YAML, "title: sysctl")
			},
		},
		{
			name: "when the behavior is internal, it serves no route",
			in:   behaviordsl.GetInput{Name: "managedfile"},
			validateFunc: func(_ behaviordsl.OpenAPIOutput, err error) {
				s.ErrorContains(err, "serves no route")
			},
		},
		{
			name: "when the spec does not build, it says so",
			in:   behaviordsl.GetInput{Name: "hostname"},
			validateFunc: func(_ behaviordsl.OpenAPIOutput, err error) {
				s.ErrorContains(err, "the spec of hostname does not build: no schema")
			},
		},
		{
			name: "when there is no such behavior, it says how to list them",
			in:   behaviordsl.GetInput{Name: "nope"},
			validateFunc: func(_ behaviordsl.OpenAPIOutput, err error) {
				s.ErrorContains(err, "osapi dsl behavior list")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(behaviordsl.OpenAPIGet(context.Background(), s.deps(), tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestCreate() {
	root := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(root, "go.mod"),
		[]byte("module github.com/osapi-io/osapi\n\ngo 1.25\n"), 0o644))
	all := filepath.Join(root, "internal", "behaviors", "all")
	s.Require().NoError(os.MkdirAll(all, 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(all, "all.go"),
		[]byte("// Package all registers every behavior.\npackage all\n\nimport (\n)\n"), 0o644))

	got, err := behaviordsl.Create(context.Background(), s.deps(), behaviordsl.CreateInput{
		Name: "neighbors", Description: "The neighbor table", Group: "network", Root: root,
	})
	s.Require().NoError(err)
	s.NotEmpty(got.Files)
	s.Contains(got.Next[1].Step, "internal/behaviors/network/neighbors")

	wd, _ := os.Getwd()
	s.Require().NoError(os.Chdir(root))
	defer func() { _ = os.Chdir(wd) }()
	got, err = behaviordsl.Create(context.Background(), s.deps(),
		behaviordsl.CreateInput{Name: "motd", Description: "The message"})
	s.Require().NoError(err, "with no root, it writes into the current directory")
	s.Contains(got.Next[1].Step, "internal/behaviors/motd")

	_, err = behaviordsl.Create(context.Background(), s.deps(),
		behaviordsl.CreateInput{Name: "Bad", Description: "x", Root: root})
	s.Error(err)

	got, err = behaviordsl.Create(context.Background(), s.deps(), behaviordsl.CreateInput{
		Name: "ntpconf", Description: "The NTP config", Requires: []string{"managedfile"},
	})
	s.Require().NoError(err, "a required behavior is wired, and its call is a step")
	s.Contains(got.Next[2].Step, "call managedfile with behavior.Call in "+
		"internal/behaviors/ntpconf/debian.go")

	_, err = behaviordsl.Create(context.Background(), s.deps(), behaviordsl.CreateInput{
		Name: "chrony", Description: "x", Requires: []string{"nosuch"},
	})
	s.ErrorContains(err, `requires "nosuch", which is not a registered behavior`)
}

func (s *MethodsPublicTestSuite) TestSpec() {
	got, err := behaviordsl.Spec(context.Background(), s.deps(), behaviordsl.SpecInput{})
	s.Require().NoError(err)
	s.NotEmpty(got.Kinds)
}

func (s *MethodsPublicTestSuite) TestValidate() {
	d := s.deps()
	d.Validate = func() []behavior.Problem { return nil }
	got, err := behaviordsl.Validate(context.Background(), d, behaviordsl.ValidateInput{})
	s.Require().NoError(err)
	s.True(got.Valid)

	d.Validate = func() []behavior.Problem {
		return []behavior.Problem{
			{Behavior: "x", Rule: "r", Message: "m", Fix: "f"},
			{Behavior: "y", Rule: "s"},
		}
	}
	got, err = behaviordsl.Validate(context.Background(), d, behaviordsl.ValidateInput{})
	s.ErrorIs(err, command.FailedError{})
	s.False(got.Valid)
	s.Equal(map[string]any{"r": "f"}, got.Fixes)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
