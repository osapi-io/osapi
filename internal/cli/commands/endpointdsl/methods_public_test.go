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

package endpointdsl_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/endpointdsl"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) deps() endpointdsl.Deps {
	return endpointdsl.Deps{
		Groups: endpoint.All,
		OpenAPI: func(g *endpoint.Group) (any, error) {
			return endpoint.OpenAPI(g, endpoint.ErrorRef)
		},
	}
}

func (s *MethodsPublicTestSuite) TestList() {
	got, err := endpointdsl.List(context.Background(), s.deps(), endpointdsl.ListInput{})
	s.Require().NoError(err)
	s.Len(got.Items, len(endpoint.All()))
	s.NotEmpty(got.Items[0].Methods)
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		in           endpointdsl.GetInput
		validateFunc func(endpointdsl.GetOutput, error)
	}{
		{
			name: "when the group is there, each method has its route and permission",
			in:   endpointdsl.GetInput{Name: "agent"},
			validateFunc: func(got endpointdsl.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("agent", got.Name)
				s.Equal(endpointdsl.Method{
					Name: "list", Verb: "GET", Path: "/api/agent",
					Permission: "agent:read", OperationID: "GetAgents",
				}, got.Methods[1])
			},
		},
		{
			name: "when no group has the name, it says to list them",
			in:   endpointdsl.GetInput{Name: "nosuch"},
			validateFunc: func(_ endpointdsl.GetOutput, err error) {
				s.ErrorContains(err, "osapi controller endpoint list")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(endpointdsl.Get(context.Background(), s.deps(), tc.in))
		})
	}
}

type chanOut = chan int

// broken has a method whose input and one whose output have no schema.
type plainIn struct {
	Name string `json:"name" validate:"required" doc:"The name"`
}

var broken = endpoint.Define("zzbroken", "Broken",
	endpoint.Get[chanOut, struct{}]("list"),
	endpoint.Get[struct{}, chanOut]("get").Sub("one"),
	endpoint.Post[plainIn, struct{}]("create"),
)

func (s *MethodsPublicTestSuite) TestMethodGet() {
	d := s.deps()
	d.Groups = func() []*endpoint.Group { return append(endpoint.All(), broken) }
	tests := []struct {
		name         string
		in           endpointdsl.MethodInput
		validateFunc func(endpointdsl.MethodOutput, error)
	}{
		{
			name: "when the method is an action, it has its route, permission, command and errors",
			in:   endpointdsl.MethodInput{Name: "agent", Method: "drain"},
			validateFunc: func(got endpointdsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Equal("POST", got.HTTPMethod)
				s.Equal("agent:write", got.Permission)
				s.True(got.AdminOnly)
				s.Equal("osapi client agent drain", got.Command)
				s.Equal("path", got.Fields[0].In)
				s.NotEmpty(got.Errors)
				s.NotEmpty(got.Examples)
				s.NotNil(got.Input)
			},
		},
		{
			name: "when a field is a query parameter, it says so, and the output has its fields",
			in:   endpointdsl.MethodInput{Name: "audit", Method: "list"},
			validateFunc: func(got endpointdsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Equal("query", got.Fields[0].In)
				s.Equal("audit:read", got.Permission)
				s.NotEmpty(got.Results)
			},
		},
		{
			name: "when the body is a struct, its fields are the body's",
			in:   endpointdsl.MethodInput{Name: "admission", Method: "admit"},
			validateFunc: func(got endpointdsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Require().Len(got.Fields, 2)
				s.Equal(endpointdsl.Field{
					Field: "machine_id", In: "path", Type: got.Fields[0].Type, Required: true,
					Constraints: got.Fields[0].Constraints, Description: got.Fields[0].Description,
				}, got.Fields[0])
				s.Equal("fingerprint", got.Fields[1].Field)
				s.Equal("body", got.Fields[1].In)
			},
		},
		{
			name: "when the method is under a sub-command, its command has it",
			in:   endpointdsl.MethodInput{Name: "facts", Method: "list"},
			validateFunc: func(got endpointdsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Equal("osapi client facts keys list", got.Command)
			},
		},
		{
			name: "when a field has no in tag, it is in the body",
			in:   endpointdsl.MethodInput{Name: "zzbroken", Method: "create"},
			validateFunc: func(got endpointdsl.MethodOutput, err error) {
				s.Require().NoError(err)
				s.Equal("body", got.Fields[0].In)
			},
		},
		{
			name: "when the input has no schema, it says so",
			in:   endpointdsl.MethodInput{Name: "zzbroken", Method: "list"},
			validateFunc: func(_ endpointdsl.MethodOutput, err error) {
				s.ErrorContains(err, "zzbroken list input")
			},
		},
		{
			name: "when the output has no schema, it says so",
			in:   endpointdsl.MethodInput{Name: "zzbroken", Method: "get"},
			validateFunc: func(_ endpointdsl.MethodOutput, err error) {
				s.ErrorContains(err, "zzbroken get output")
			},
		},
		{
			name: "when the group has no such method, it says how to list them",
			in:   endpointdsl.MethodInput{Name: "agent", Method: "nope"},
			validateFunc: func(_ endpointdsl.MethodOutput, err error) {
				s.ErrorContains(err, "osapi controller endpoint get --name agent")
			},
		},
		{
			name: "when there is no such group, it says how to list them",
			in:   endpointdsl.MethodInput{Name: "nope", Method: "list"},
			validateFunc: func(_ endpointdsl.MethodOutput, err error) {
				s.ErrorContains(err, "osapi controller endpoint list")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(endpointdsl.MethodGet(context.Background(), d, tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestOpenAPIGet() {
	tests := []struct {
		name         string
		in           endpointdsl.GetInput
		openapiErr   error
		validateFunc func(endpointdsl.OpenAPIOutput, error)
	}{
		{
			name: "when the spec builds, it is YAML",
			in:   endpointdsl.GetInput{Name: "agent"},
			validateFunc: func(got endpointdsl.OpenAPIOutput, err error) {
				s.Require().NoError(err)
				s.Contains(got.YAML, "/api/agent")
				s.Contains(got.YAML, "  schemas:\n", "it is indented by two")
			},
		},
		{
			name:       "when the spec does not build, it is the error",
			in:         endpointdsl.GetInput{Name: "agent"},
			openapiErr: errors.New("bad schema"),
			validateFunc: func(_ endpointdsl.OpenAPIOutput, err error) {
				s.EqualError(err, "the spec of agent does not build: bad schema")
			},
		},
		{
			name: "when no group has the name, it is the error",
			in:   endpointdsl.GetInput{Name: "nosuch"},
			validateFunc: func(_ endpointdsl.OpenAPIOutput, err error) {
				s.ErrorContains(err, "no endpoint group")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := s.deps()
			if tc.openapiErr != nil {
				d.OpenAPI = func(*endpoint.Group) (any, error) { return nil, tc.openapiErr }
			}
			tc.validateFunc(endpointdsl.OpenAPIGet(context.Background(), d, tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestCreate() {
	dir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module github.com/osapi-io/osapi\n\ngo 1.25\n"), 0o644))
	all := filepath.Join(dir, "internal", "controller", "endpoints", "all")
	s.Require().NoError(os.MkdirAll(all, 0o755))
	s.Require().NoError(os.WriteFile(
		filepath.Join(all, "all.go"),
		[]byte(
			"// Package all registers every endpoint group.\npackage all\n\nimport (\n)\n",
		),
		0o644,
	))
	d := s.deps()
	d.Dir = dir

	got, err := endpointdsl.Create(context.Background(), d,
		endpointdsl.CreateInput{Name: "gizmo", Description: "Gizmos"})
	s.Require().NoError(err)
	s.Equal("gizmo", got.Name)
	s.NotEmpty(got.Files)
	s.Contains(
		got.Next,
		endpointdsl.Step{Command: "go test ./internal/controller/endpoints/gizmo/..."},
	)

	_, err = endpointdsl.Create(context.Background(), d,
		endpointdsl.CreateInput{Name: "Bad", Description: "x"})
	s.Error(err)
}

func (s *MethodsPublicTestSuite) TestSpec() {
	got, err := endpointdsl.Spec(context.Background(), s.deps(), endpointdsl.SpecInput{})
	s.Require().NoError(err)
	s.Equal(endpoint.Spec.Sections(), got.Kinds)
}

func (s *MethodsPublicTestSuite) TestValidate() {
	got, err := endpointdsl.Validate(context.Background(), s.deps(), endpointdsl.ValidateInput{})
	s.Require().NoError(err)
	s.True(got.Valid)
	s.Equal(len(endpoint.All()), got.Groups)

	broken := endpointdsl.Deps{Groups: func() []*endpoint.Group {
		return []*endpoint.Group{endpoint.Define("Bad", "Bad")}
	}}
	got, err = endpointdsl.Validate(context.Background(), broken, endpointdsl.ValidateInput{})
	s.ErrorIs(err, command.FailedError{})
	s.False(got.Valid)
	s.NotEmpty(got.Fixes)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
