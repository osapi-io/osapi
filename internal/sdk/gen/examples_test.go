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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/sdk/gen/sdkmodel"
)

type ExamplesTestSuite struct {
	genSuite
}

func (s *ExamplesTestSuite) readExample(
	name string,
) string {
	data, err := os.ReadFile(filepath.Join(s.root, exampleDir, name))
	s.Require().NoError(err)

	return string(data)
}

func (s *ExamplesTestSuite) TestRun() {
	// A program of the last run, which a service that is gone leaves.
	s.write(exampleDir+"/gone.go", header+"\n\npackage main\n")
	s.write(exampleDir+"/go.mod", "module examples\n")
	s.Require().NoError(generate(s.root, false))

	tests := []struct {
		name         string
		validateFunc func()
	}{
		{
			name: "when a node's method takes the target and an input",
			validateFunc: func() {
				dns := s.readExample("network_dns.go")
				s.Contains(dns, header)
				s.Contains(dns, "c.NetworkDNS.Get(ctx, target, \"eth0\")")
				s.Contains(dns, "client.NetworkDNSUpdateInput{\n\t\tInterfaceName: \"eth0\",\n"+
					"\t\tOverrideDHCP:  true,\n\t}")
				s.Contains(dns, `{"update", "Override DHCP", runUpdate}`)
				s.Contains(dns, `target := os.Getenv("OSAPI_TARGET")`)
				s.Contains(dns, "return show(resp.Data)")
			},
		},
		{
			name: "when a method of the controller takes its keys and its query",
			validateFunc: func() {
				job := s.readExample("job.go")
				s.Contains(job, "c.Job.List(ctx, client.JobListParams{})")
				s.Contains(job, "client.JobListParams{\n\t\tAll:   true,\n\t\tLimit: 3,\n\t}")
				s.Contains(job, `{"list-2", "Three jobs", runList2}`)
				s.Contains(job, `c.Job.Delete(ctx, "550e8400-e29b-41d4-a716-446655440000")`)
				s.Contains(job, "fmt.Println(\"done\")")
				s.NotContains(job, "OSAPI_TARGET")
				s.Contains(
					s.readExample("node_file.go"),
					`c.NodeFile.Status(ctx, target, "/etc/hosts")`,
				)
			},
		},
		{
			name: "when a method sends a form, the program reads the file and calls it",
			validateFunc: func() {
				file := s.readExample("file.go")
				s.Contains(file, `content, err := os.ReadFile("/tmp/a.conf")`)
				s.Contains(file, `c.File.Upload(ctx, "a.conf", "raw", bytes.NewReader(content))`)
			},
		},
		{
			name: "when a service of the last run is gone, its program goes",
			validateFunc: func() {
				s.NoFileExists(filepath.Join(s.root, exampleDir, "gone.go"))
				s.FileExists(filepath.Join(s.root, exampleDir, "go.mod"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc()
		})
	}
}

func (s *ExamplesTestSuite) TestRunNoExample() {
	s.write(exampleDir+"/file.go", header+"\n\npackage main\n")
	s.edit("      x-osapi-examples:\n        - doc: Upload a file\n"+
		"          input: {name: a.conf, file: /tmp/a.conf}\n", "")

	s.Require().NoError(generate(s.root, false))

	s.NoFileExists(filepath.Join(s.root, exampleDir, "file.go"),
		"a service with no example has no program, and the old one goes")
}

func (s *ExamplesTestSuite) TestRunErrors() {
	tests := []struct {
		name    string
		setup   func()
		wantErr string
	}{
		{
			name:    "when a person wrote the program of a service",
			setup:   func() { s.write(exampleDir+"/job.go", "package main\n") },
			wantErr: "examples/sdk/client/job.go is a person's",
		},
		{
			name:    "when the directory of the examples is a file",
			setup:   func() { s.write("examples/sdk/client", "") },
			wantErr: "not a directory",
		},
		{
			name:    "when the SDK's package is not there",
			setup:   func() { s.Require().NoError(os.RemoveAll(filepath.Join(s.root, dir))) },
			wantErr: "no such file or directory",
		},
		{
			name: "when a name of the spec is not a name of Go",
			setup: func() {
				path := filepath.Join(s.root, spec)
				data, err := os.ReadFile(path)
				s.Require().NoError(err)
				s.write(spec, strings.ReplaceAll(string(data), "DnsEntry", "Dns-Entry"))
			},
			wantErr: "network_dns",
		},
		{
			name: "when an example has no value for a key",
			setup: func() {
				s.edit("input: {id: 550e8400-e29b-41d4-a716-446655440000}", "input: {}")
			},
			wantErr: `Job.Delete: example "A job": it has no id`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			s.ErrorContains(generate(s.root, false), tc.wantErr)
		})
	}
}

// edit changes the spec under the root, old to repl.
func (s *ExamplesTestSuite) edit(
	old string,
	repl string,
) {
	path := filepath.Join(s.root, spec)
	data, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Require().Contains(string(data), old)
	s.write(spec, replaceOnce(string(data), old, repl))
}

func (s *ExamplesTestSuite) TestCallCode() {
	str := &sdkmodel.TypeRef{Kind: sdkmodel.Scalar, Name: "string"}
	agent := &sdkmodel.Service{Name: "Agent"}
	get := &sdkmodel.Method{
		Name:       "Get",
		PathParams: []*sdkmodel.Param{{Name: "hostname", ArgName: "hostname", Type: str}},
	}
	limit := &sdkmodel.Param{
		Name:   "limit",
		GoName: "Limit",
		Type:   &sdkmodel.TypeRef{Kind: sdkmodel.Scalar, Name: "int"},
	}
	list := &sdkmodel.Method{Name: "List", Query: []*sdkmodel.Param{limit}}
	tags := &sdkmodel.Field{
		Name:   "tags",
		GoName: "Tags",
		Type:   &sdkmodel.TypeRef{Kind: sdkmodel.Slice, Elem: str},
	}
	create := &sdkmodel.Method{
		Name:  "Create",
		Input: &sdkmodel.Type{Name: "AgentCreateInput", Fields: []*sdkmodel.Field{tags}},
	}
	form := func(fields ...*sdkmodel.Field) *sdkmodel.Method {
		return &sdkmodel.Method{Name: "Upload", Multipart: true, Form: fields}
	}
	name := &sdkmodel.Field{Name: "name", Type: str, Required: true}
	kind := &sdkmodel.Field{Name: "kind", Type: str, Default: "raw"}
	force := &sdkmodel.Field{
		Name: "force",
		Type: &sdkmodel.TypeRef{Kind: sdkmodel.Scalar, Name: "bool"},
	}
	file := &sdkmodel.Field{Name: "file", Type: str, File: true}
	read := "content, err := os.ReadFile(\"/tmp/a\")\n" +
		"if err != nil {\n\treturn err\n}\n\n"

	tests := []struct {
		name    string
		method  *sdkmodel.Method
		input   map[string]any
		want    string
		wantErr string
	}{
		{
			name:   "when the controller's method has a host in its path, it is the example's",
			method: get,
			input:  map[string]any{"hostname": "web-01"},
			want:   `c.Agent.Get(ctx, "web-01")`,
		},
		{
			name:    "when the example has no host for the path",
			method:  get,
			input:   map[string]any{},
			wantErr: "it has no hostname",
		},
		{
			name:    "when a value of the query is not its type",
			method:  list,
			input:   map[string]any{"limit": "many"},
			wantErr: "limit: many is not a int",
		},
		{
			name:    "when a field of the input is not its type",
			method:  create,
			input:   map[string]any{"tags": "web"},
			wantErr: "tags: web is not a list",
		},
		{
			name:   "when a form method has a file, the call reads it first",
			method: form(force, kind, name, file),
			input:  map[string]any{"name": "a", "file": "/tmp/a"},
			want: read +
				`c.Agent.Upload(ctx, "a", false, "raw", bytes.NewReader(content))`,
		},
		{
			name:    "when the form has no file",
			method:  form(name),
			input:   map[string]any{"name": "a"},
			wantErr: "the form has 0 files, not one",
		},
		{
			name:    "when the example has no path of the file",
			method:  form(file),
			input:   map[string]any{},
			wantErr: "it has no path of file",
		},
		{
			name:    "when the example has no required field of the form",
			method:  form(name, file),
			input:   map[string]any{"file": "/tmp/a"},
			wantErr: "it has no name",
		},
		{
			name:    "when a field of the form is not its type",
			method:  form(force, file),
			input:   map[string]any{"force": "yes", "file": "/tmp/a"},
			wantErr: "force: yes is not a bool",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			code, pre, err := callCode(agent, tc.method, &sdkmodel.Example{Input: tc.input})

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			text := ""
			for _, p := range pre {
				text += strings.TrimSpace(fmt.Sprintf("%#v", p)) + "\n"
			}
			s.Equal(tc.want, text+fmt.Sprintf("%#v", code))
		})
	}
}

func (s *ExamplesTestSuite) TestZeroJSON() {
	scalar := func(name string) *sdkmodel.TypeRef {
		return &sdkmodel.TypeRef{Kind: sdkmodel.Scalar, Name: name}
	}
	tests := []struct {
		name string
		typ  *sdkmodel.TypeRef
		want any
	}{
		{name: "when it is a string", typ: scalar("string"), want: ""},
		{name: "when it is a bool", typ: scalar("bool"), want: false},
		{name: "when it is a number", typ: scalar("int"), want: float64(0)},
		{name: "when it is not a scalar", typ: &sdkmodel.TypeRef{Kind: sdkmodel.Slice}, want: nil},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, zeroJSON(tc.typ))
		})
	}
}

func (s *ExamplesTestSuite) TestValue() {
	scalar := func(name string) *sdkmodel.TypeRef {
		return &sdkmodel.TypeRef{Kind: sdkmodel.Scalar, Name: name}
	}
	named := &sdkmodel.TypeRef{Kind: sdkmodel.Named, Name: "Entry"}

	tests := []struct {
		name    string
		typ     *sdkmodel.TypeRef
		value   any
		want    string
		wantErr string
	}{
		{name: "when it is a string", typ: scalar("string"), value: "a", want: `"a"`},
		{name: "when it is a bool", typ: scalar("bool"), value: true, want: "true"},
		{name: "when it is an int64", typ: scalar("int64"), value: float64(3), want: "3"},
		{name: "when it is a float64", typ: scalar("float64"), value: 0.5, want: "0.5"},
		{
			name:  "when it is a list",
			typ:   &sdkmodel.TypeRef{Kind: sdkmodel.Slice, Elem: scalar("int")},
			value: []any{float64(1), float64(2)},
			want:  "[]int{1, 2}",
		},
		{
			name:  "when it is a map",
			typ:   &sdkmodel.TypeRef{Kind: sdkmodel.Map, Elem: scalar("string")},
			value: map[string]any{"b": "2", "a": "1"},
			want:  "map[string]string{\n\t\"a\": \"1\",\n\t\"b\": \"2\",\n}",
		},
		{
			name:    "when an int has a fraction",
			typ:     scalar("int"),
			value:   1.5,
			wantErr: "1.5 is not a int",
		},
		{
			name:    "when a time has no literal",
			typ:     scalar("time.Time"),
			value:   "2026-01-01T00:00:00Z",
			wantErr: "is not a time.Time",
		},
		{
			name:    "when a named type has no literal",
			typ:     named,
			value:   map[string]any{},
			wantErr: "a value of Entry has no literal",
		},
		{
			name:    "when a list's items have no literal",
			typ:     &sdkmodel.TypeRef{Kind: sdkmodel.Slice, Elem: named},
			value:   []any{},
			wantErr: "a value of Entry has no literal",
		},
		{
			name:    "when an item of a list is not its type",
			typ:     &sdkmodel.TypeRef{Kind: sdkmodel.Slice, Elem: scalar("bool")},
			value:   []any{"yes"},
			wantErr: "yes is not a bool",
		},
		{
			name:    "when a map is not a map",
			typ:     &sdkmodel.TypeRef{Kind: sdkmodel.Map, Elem: scalar("string")},
			value:   "a=b",
			wantErr: "a=b is not a map",
		},
		{
			name:    "when a map's values have no literal",
			typ:     &sdkmodel.TypeRef{Kind: sdkmodel.Map, Elem: named},
			value:   map[string]any{},
			wantErr: "a value of Entry has no literal",
		},
		{
			name:    "when a value of a map is not its type",
			typ:     &sdkmodel.TypeRef{Kind: sdkmodel.Map, Elem: scalar("float64")},
			value:   map[string]any{"a": "b"},
			wantErr: "b is not a float64",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			code, err := value(tc.typ, tc.value)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, fmt.Sprintf("%#v", code))
		})
	}
}

func TestExamplesTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ExamplesTestSuite))
}
