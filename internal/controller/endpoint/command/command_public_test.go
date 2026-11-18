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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/reference"
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

type formInput struct {
	Name string `json:"name" in:"form"`
	Note string `json:"note" in:"form"`
	File []byte `json:"file" in:"form"`
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
	endpoint.Post[formInput, thing]("upload").Sub("upload"),
	endpoint.Get[getInput, thing]("probe").Sub("probe").
		Answers(http.StatusServiceUnavailable, "Not ready."),
)

// caller records the request, and answers with body or err.
type caller struct {
	method, path string
	contentType  string
	body         []byte
	answer       []byte
	err          error
}

func (c *caller) Send(
	_ context.Context,
	method string,
	path string,
	contentType string,
	body []byte,
) (*client.Response[json.RawMessage], error) {
	c.method, c.path, c.contentType, c.body = method, path, contentType, body
	if c.err != nil {
		return nil, c.err
	}

	return client.NewResponse(json.RawMessage(c.answer), c.answer), nil
}

type CommandPublicTestSuite struct {
	suite.Suite
}

func (s *CommandPublicTestSuite) TestReadFile() {
	var reported error
	parent := &cobra.Command{Use: "client"}
	s.Require().NoError(command.Attach(parent, command.Deps{
		Caller: func() command.Caller { return &caller{} },
		JSON:   func() bool { return false },
		Error:  func(err error) { reported = err },
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

func (s *CommandPublicTestSuite) TestBuild() {
	tests := []struct {
		name     string
		method   string
		in       any
		wantPath string
		wantBody string
		wantForm map[string]string
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
			name:     "when the input has form fields, the body is a form without its zero fields",
			method:   "upload",
			in:       formInput{Name: "a.conf", File: []byte("data")},
			wantPath: "/api/thing/upload",
			wantForm: map[string]string{"name": "a.conf", "file": "data"},
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
			path, body, contentType, err := command.Build(group, group.Method(tc.method), tc.in)

			if tc.wantErr {
				s.Error(err)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantPath, path)
			if tc.wantForm != nil {
				_, params, err := mime.ParseMediaType(contentType)
				s.Require().NoError(err)
				form, err := multipart.NewReader(bytes.NewReader(body), params["boundary"]).
					ReadForm(1 << 20)
				s.Require().NoError(err)
				s.Equal([]string{tc.wantForm["name"]}, form.Value["name"])
				s.NotContains(form.Value, "note")
				f, err := form.File["file"][0].Open()
				s.Require().NoError(err)
				data, _ := io.ReadAll(f)
				s.Equal(tc.wantForm["file"], string(data))

				return
			}
			s.Equal("application/json", contentType)
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
	acted := ""
	probed := ""
	command.Render(group, "probe", func(_ *cobra.Command, _ getInput, out thing) {
		probed = out.Name
	})
	command.Acts(group, "retry", func(_ *cobra.Command, in retryInput, out thing) {
		acted = in.Body.Reason + ":" + out.Name
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
			name:   "when the method acts on its answer, it runs with --json too",
			method: "retry",
			in:     retryInput{Body: Retry{Reason: "flaky"}},
			json:   true,
			caller: &caller{answer: []byte(`{"name":"web-03"}`)},
			validateFunc: func(_ *caller, out string, err error) {
				s.Require().NoError(err)
				s.Equal("flaky:web-03", acted)
				s.Empty(out)
			},
		},
		{
			name:   "when the method answers a status with its output, it renders that output",
			method: "probe",
			in:     getInput{Name: "web"},
			caller: &caller{err: &client.APIError{
				StatusCode: http.StatusServiceUnavailable,
				Body:       []byte(`{"name":"not ready"}`),
			}},
			validateFunc: func(_ *caller, _ string, err error) {
				s.Require().NoError(err)
				s.Equal("not ready", probed)
			},
		},
		{
			name:   "when a status is not one the method answers, it is the error",
			method: "probe",
			in:     getInput{Name: "web"},
			caller: &caller{err: &client.APIError{StatusCode: http.StatusNotFound}},
			validateFunc: func(_ *caller, _ string, err error) {
				s.Error(err)
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
				case retryInput:
					command.Run(cmd, group, tc.method, in)
				case badBody:
					command.Run(cmd, group, tc.method, in)
				}
			})

			tc.validateFunc(tc.caller, out, gotErr)
			// The fake answers with no traceresponse header, so the call
			// has no trace for a renderer to print.
			s.Empty(command.TraceID())
		})
	}
}

// TestAttachUnderPrefix puts a group under the commands of its prefix, made
// when the parent has none and found when it has, with its notes.
func (s *CommandPublicTestSuite) TestAttachUnderPrefix() {
	g := endpoint.Define("zzbehavior", "A behavior",
		endpoint.Prefix("zznext/zznode/{hostname}"),
		endpoint.CommandNotes("See its page."),
		endpoint.Get[getInput, thing]("get"))
	get := &cobra.Command{Use: "get"}
	command.Register(g, get)

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
