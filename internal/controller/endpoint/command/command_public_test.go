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
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type getInput struct {
	Name  string `json:"name"  in:"path"`
	Limit int    `json:"limit" in:"query"`
	Skip  string `json:"-"`
	note  string //nolint:unused
}

type setInput struct {
	Name  string `json:"name"  in:"path"`
	Value string `json:"value"`
	Count int
}

type Retry struct {
	Reason string `json:"reason"`
}

type retryInput struct {
	Body Retry `in:"body"`
}

type badBody struct {
	Body func() `in:"body"`
}

type thing struct {
	Name string `json:"name"`
}

var group = endpoint.Define("thing", "Things",
	endpoint.Get[getInput, thing]("get"),
	endpoint.Put[setInput, thing]("set"),
	endpoint.Post[retryInput, thing]("retry").Sub("retry"),
	endpoint.Post[badBody, thing]("bad").Sub("bad"),
	endpoint.Get[string, thing]("raw").Sub("raw"),
)

// caller records the request, and answers with body or err.
type caller struct {
	method, path string
	body         []byte
	answer       []byte
	err          error
}

func (c *caller) Request(
	_ context.Context,
	method string,
	path string,
	body []byte,
) (*client.Response[json.RawMessage], error) {
	c.method, c.path, c.body = method, path, body
	if c.err != nil {
		return nil, c.err
	}

	return client.NewResponse(json.RawMessage(c.answer), c.answer), nil
}

type CommandPublicTestSuite struct {
	suite.Suite
}

func (s *CommandPublicTestSuite) TestBuild() {
	tests := []struct {
		name     string
		method   string
		in       any
		wantPath string
		wantBody string
		wantErr  bool
	}{
		{
			name:     "when the input has a key and a query, they go in the path",
			method:   "get",
			in:       getInput{Name: "web 01", Limit: 3},
			wantPath: "/api/thing/web%2001?limit=3",
		},
		{
			name:     "when a query field is zero, it is left out",
			method:   "get",
			in:       getInput{Name: "web"},
			wantPath: "/api/thing/web",
		},
		{
			name:     "when the input has other fields, they are the JSON body",
			method:   "set",
			in:       setInput{Name: "web", Value: "v", Count: 2},
			wantPath: "/api/thing/web",
			wantBody: `{"Count":2,"value":"v"}`,
		},
		{
			name:     "when a field is the whole body, it is the body",
			method:   "retry",
			in:       retryInput{Body: Retry{Reason: "again"}},
			wantPath: "/api/thing/retry",
			wantBody: `{"reason":"again"}`,
		},
		{
			name:     "when the input is not a struct, there is no body",
			method:   "raw",
			in:       "x",
			wantPath: "/api/thing/raw",
		},
		{
			name:    "when the body cannot be encoded, it is an error",
			method:  "bad",
			in:      badBody{Body: func() {}},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			path, body, err := command.Build(group, group.Method(tc.method), tc.in)

			if tc.wantErr {
				s.Error(err)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantPath, path)
			if tc.wantBody == "" {
				s.Nil(body)
			} else {
				s.JSONEq(tc.wantBody, string(body))
			}
		})
	}
}

// capture is what f prints to stdout.
func capture(
	f func(),
) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = stdout
	out, _ := io.ReadAll(r)

	return string(out)
}

func (s *CommandPublicTestSuite) TestRun() {
	rendered := ""
	command.Render(group, "get", func(_ *cobra.Command, in getInput, out thing) {
		rendered = in.Name + ":" + out.Name
	})

	tests := []struct {
		name         string
		method       string
		in           any
		json         bool
		caller       *caller
		validateFunc func(c *caller, out string, err error)
	}{
		{
			name:   "when the method has a renderer, it gets the input and the output",
			method: "get",
			in:     getInput{Name: "web"},
			caller: &caller{answer: []byte(`{"name":"web-01"}`)},
			validateFunc: func(c *caller, _ string, err error) {
				s.Require().NoError(err)
				s.Equal("GET", c.method)
				s.Equal("web:web-01", rendered)
			},
		},
		{
			name:   "when --json is given, the body prints as it came",
			method: "get",
			in:     getInput{Name: "web"},
			json:   true,
			caller: &caller{answer: []byte(`{"name":"web-01"}`)},
			validateFunc: func(_ *caller, out string, err error) {
				s.Require().NoError(err)
				s.Contains(out, `"web-01"`)
			},
		},
		{
			name:   "when the method has no renderer, the body prints as JSON",
			method: "set",
			in:     setInput{Name: "web"},
			caller: &caller{answer: []byte(`{"name":"web-02"}`)},
			validateFunc: func(c *caller, out string, err error) {
				s.Require().NoError(err)
				s.Equal("PUT", c.method)
				s.Contains(out, `"web-02"`)
			},
		},
		{
			name:   "when the call fails, the error is reported",
			method: "get",
			in:     getInput{Name: "web"},
			caller: &caller{err: errors.New("api error (404): not found")},
			validateFunc: func(_ *caller, _ string, err error) {
				s.EqualError(err, "api error (404): not found")
			},
		},
		{
			name:   "when the answer does not decode, the error is reported",
			method: "get",
			in:     getInput{Name: "web"},
			caller: &caller{answer: []byte(`{"name":3}`)},
			validateFunc: func(_ *caller, _ string, err error) {
				s.ErrorContains(err, "decode thing get")
			},
		},
		{
			name:   "when the answer is empty, the renderer gets a zero output",
			method: "get",
			in:     getInput{Name: "web"},
			caller: &caller{answer: nil},
			validateFunc: func(_ *caller, _ string, err error) {
				s.Require().NoError(err)
				s.Equal("web:", rendered)
			},
		},
		{
			name:   "when the request cannot be built, the error is reported",
			method: "bad",
			in:     badBody{Body: func() {}},
			caller: &caller{},
			validateFunc: func(_ *caller, _ string, err error) {
				s.Error(err)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var gotErr error
			parent := &cobra.Command{Use: "client"}
			s.Require().NoError(command.Attach(parent, command.Deps{
				Caller: func() command.Caller { return tc.caller },
				JSON:   func() bool { return tc.json },
				Error:  func(err error) { gotErr = err },
			}))
			cmd := &cobra.Command{}
			cmd.SetContext(context.Background())

			out := capture(func() {
				switch in := tc.in.(type) {
				case getInput:
					command.Run(cmd, group, tc.method, in)
				case setInput:
					command.Run(cmd, group, tc.method, in)
				case badBody:
					command.Run(cmd, group, tc.method, in)
				}
			})

			tc.validateFunc(tc.caller, out, gotErr)
		})
	}
}

func (s *CommandPublicTestSuite) TestAttach() {
	g := endpoint.Define("zzattach", "Attached things", endpoint.Get[getInput, thing]("get"))
	get := &cobra.Command{Use: "get"}
	command.Register(g, get)

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
	s.Contains(made.Commands(), get)

	// A parent that has the group, and a command of the same name, is a
	// command written twice.
	parent = &cobra.Command{Use: "client"}
	existing := &cobra.Command{Use: "zzattach"}
	existing.AddCommand(&cobra.Command{Use: "get"})
	parent.AddCommand(existing)
	err := command.Attach(parent, command.Deps{})
	s.Require().Error(err)
	s.True(strings.Contains(err.Error(), "both a hand-written command"))

	// A command with a Sub annotation goes under that command: a new one
	// when the group has none, the one there when it has.
	g2 := endpoint.Define("zznested", "Nested things", endpoint.Get[getInput, thing]("get"))
	nested := &cobra.Command{Use: "get", Annotations: map[string]string{command.SubKey: "trace"}}
	nested2 := &cobra.Command{Use: "list", Annotations: map[string]string{command.SubKey: "trace"}}
	command.Register(g2, nested, nested2)
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
	group := &cobra.Command{Use: "zznested"}
	namedTrace := &cobra.Command{Use: "trace", Short: "Read a trace"}
	s.Require().NoError(command.Attach(parent, command.Deps{}, group, namedTrace))
	s.Len(namedTrace.Commands(), 2)
	for _, c := range group.Commands() {
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

func TestCommandPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CommandPublicTestSuite))
}
