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

package commandgen_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	commandgen "github.com/osapi-io/osapi/internal/cli/command/gen"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

type OutputsPublicTestSuite struct {
	suite.Suite
}

type noIn struct{}

type node struct {
	Name string `json:"name" doc:"The name."`
	Seen string `json:"seen" doc:"When it was seen." render:"age"`
}

// ansi matches the escape codes of a color, which the test leaves out.
var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func (s *OutputsPublicTestSuite) TestOutputs() {
	tests := []struct {
		name  string
		group *endpoint.Group
		want  []string
	}{
		{
			name: "when an example has an output, the CLI's screen and the JSON",
			group: endpoint.Define("zzoutputs", "Outputs",
				endpoint.Get[noIn, node]("get").
					Example("--json").
					Example("", "A node").
					Prints(node{Name: "web-01", Seen: "2026-01-15T09:59:30Z"}),
			),
			want: []string{
				"A node:\n\n```ansi\n$ osapi client zzoutputs get\n",
				"web-01",
				"30s",
				"trace " + commandgen.ExampleTraceID,
				"With `--json`, the API's response:\n\n```json\n{\n  \"name\": \"web-01\",",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			m := tc.group.Methods[0]
			got := ansi.ReplaceAllString(
				commandgen.Outputs(tc.group, m, "osapi client zzoutputs get"),
				"",
			)
			for _, w := range tc.want {
				s.Contains(got, w)
			}
		})
	}
}

func (s *OutputsPublicTestSuite) TestOutputsLocal() {
	g := command.Define("zzlocalout", "Local",
		command.Method[noIn, node]("get").Example("").Prints(node{Name: "web-01"}))
	got := ansi.ReplaceAllString(commandgen.Outputs(g, g.Methods[0], "osapi zzlocalout get"), "")
	s.Contains(got, "web-01")
	s.Contains(got, "With `--json`, the command prints:")
	s.NotContains(got, "trace", "a local command calls no API, so it has no trace")
}

func (s *OutputsPublicTestSuite) TestOutputsNoBody() {
	type drainIn struct {
		Hostname string `json:"hostname" in:"path"`
	}
	g := endpoint.Define("zznobody", "No body",
		endpoint.Post[drainIn, struct{}]("drain").
			Says("draining").
			Example("--hostname web-01").
			Example("--hostname web-01 --json"),
	)
	got := ansi.ReplaceAllString(
		commandgen.Outputs(g, g.Methods[0], "osapi client zznobody drain"),
		"",
	)
	s.Contains(got, "$ osapi client zznobody drain --hostname web-01\n")
	s.Contains(got, "web-01 draining")
	s.Contains(got, "trace "+commandgen.ExampleTraceID)
	s.Contains(
		got,
		"With `--json`, the command prints:\n\n```json\n{\"hostname\":\"web-01\",\"result\":\"draining\"}",
	)
	s.NotContains(
		got,
		"--hostname web-01 --json",
		"an example with --json shows no output of its own",
	)
}

func (s *OutputsPublicTestSuite) TestOutputsNone() {
	g := endpoint.Define("zznone", "None", endpoint.Get[noIn, node]("get").Example(""))
	s.Empty(commandgen.Outputs(g, g.Methods[0], "osapi client zznone get"))
}

func (s *OutputsPublicTestSuite) TestOutputsDoc() {
	g := endpoint.Define("zzdoc", "Doc",
		endpoint.Get[noIn, node]("get").Example("").Prints(node{Name: "web-01"}))
	s.Contains(commandgen.Outputs(g, g.Methods[0], "osapi client zzdoc get"), "The command prints:")
}

func TestOutputsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OutputsPublicTestSuite))
}
