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

package endpoint_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

type ExampleInputPublicTestSuite struct {
	suite.Suite
}

// specExampleIn has a flag of each kind that an example can set, a path key
// with a short flag, and a body that is a struct.
type specExampleIn struct {
	Hostname string            `json:"hostname" in:"path"  flag:"target" short:"T"`
	Limit    int               `json:"limit"    in:"query"`
	Ratio    float64           `json:"ratio"    in:"query"`
	Force    *bool             `json:"force"    in:"query"`
	Body     specExampleBody   `                in:"body"`
	Labels   map[string]string `json:"labels"`
}

type specExampleBody struct {
	Servers []string `json:"servers"`
	Note    string   `json:"note"`
}

type specExampleForm struct {
	Name string `json:"name" in:"form"`
	File []byte `json:"file" in:"form"`
}

// examples are the examples of the operation of the method m of a group.
func (s *ExampleInputPublicTestSuite) examples(
	m *endpoint.Method,
) (any, error) {
	doc, err := endpoint.OpenAPI(endpoint.Define("node", "The nodes", m))
	if err != nil {
		return nil, err
	}
	for _, path := range doc.Paths.InMatchingOrder() {
		for _, op := range doc.Paths.Value(path).Operations() {
			return op.Extensions[endpoint.ExamplesExtension], nil
		}
	}

	return nil, nil
}

func (s *ExampleInputPublicTestSuite) TestExamples() {
	tests := []struct {
		name   string
		method *endpoint.Method
		want   any
	}{
		{
			name: "when each kind of flag sets its parameter or property",
			method: endpoint.Put[specExampleIn, Node]("update").
				Describe("Update a node").
				Example(`--target web-01 --limit 3 --ratio=0.5 --force --servers 1.1.1.1 `+
					`--servers 9.9.9.9 --note "a b" --labels a=b --labels c=d=e --json`,
					"Every flag").
				Example("-T web-01").
				Example("-T web-01 --json").
				Example("-T web-02 --nope 1 --typo=2 --gone --json"),
			want: []endpoint.SpecExample{
				{Doc: "Every flag", Input: map[string]any{
					"hostname": "web-01",
					"limit":    int64(3),
					"ratio":    0.5,
					"force":    true,
					"servers":  []any{"1.1.1.1", "9.9.9.9"},
					"note":     "a b",
					"labels":   map[string]any{"a": "b", "c": "d=e"},
				}},
				{Doc: "Update a node, with -T web-01", Input: map[string]any{"hostname": "web-01"}},
				{
					Doc:   "Update a node, with -T web-02 --nope 1 --typo=2 --gone --json",
					Input: map[string]any{"hostname": "web-02"},
				},
			},
		},
		{
			name:   "when the first example has no doc, it is the summary",
			method: endpoint.Get[noInput, Node]("list").Describe("List the nodes").Example(""),
			want:   []endpoint.SpecExample{{Doc: "List the nodes", Input: map[string]any{}}},
		},
		{
			name:   "when a method has no example, the operation has none",
			method: endpoint.Get[noInput, Node]("list"),
			want:   nil,
		},
		{
			name: "when the body is a multipart form, the SDK has no method to call",
			method: endpoint.Post[specExampleForm, Node](
				"upload",
			).Example("--name a --file /tmp/a"),
			want: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := s.examples(tc.method)

			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *ExampleInputPublicTestSuite) TestExamplesErrors() {
	tests := []struct {
		name    string
		flags   string
		wantErr string
	}{
		{
			name:    "when a word is not a flag",
			flags:   "web-01",
			wantErr: "web-01 is not a flag",
		},
		{
			name:    "when the last flag has no value",
			flags:   "--limit",
			wantErr: "--limit has no value",
		},
		{
			name:    "when a number does not parse",
			flags:   "--limit many",
			wantErr: "--limit: strconv.ParseInt",
		},
		{
			name:    "when a bool does not parse",
			flags:   "--force=maybe",
			wantErr: "--force=maybe: strconv.ParseBool",
		},
		{
			name:    "when an item of a list does not parse",
			flags:   "--counts x",
			wantErr: "--counts: strconv.ParseInt",
		},
		{
			name:    "when a map's value is not key=value",
			flags:   "--labels a",
			wantErr: "a is not key=value",
		},
		{
			name:    "when a quote does not close",
			flags:   `--note "a b`,
			wantErr: "has no closing quote",
		},
		{
			name:    "when a quoted word is not a Go string",
			flags:   `--note "\q"`,
			wantErr: "invalid syntax",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, err := s.examples(endpoint.Put[specExampleBad, Node]("update").Example(tc.flags))

			s.ErrorContains(err, tc.wantErr)
		})
	}
}

// specExampleBad has the flags that the cases of TestExamplesErrors break.
type specExampleBad struct {
	Limit  int               `json:"limit"  in:"query"`
	Force  bool              `json:"force"  in:"query"`
	Counts []int             `json:"counts"`
	Labels map[string]string `json:"labels"`
	Note   string            `json:"note"`
}

func TestExampleInputPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ExampleInputPublicTestSuite))
}
