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

package specgen_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoint/gen/specgen"
)

type SpecgenPublicTestSuite struct {
	suite.Suite
}

type node struct {
	Name string  `json:"name" doc:"The name." example:"web-01"`
	Size int64   `json:"size" doc:"The size." example:"134217728"`
	Load float64 `json:"load" doc:"The load." example:"2.5"`
}

type other struct {
	Name string `json:"name" doc:"Another name."`
}

type listInput struct{}

// spec is the spec of a group named name with one list of out, under title.
func spec[Out any](
	name string,
	title string,
) *openapi3.T {
	doc, err := endpoint.OpenAPI(endpoint.Define(name, "The "+name,
		endpoint.Title(title),
		endpoint.Tag(name+"_operations", name, "The "+name+"."),
		endpoint.Get[listInput, Out]("list"),
	))
	if err != nil {
		panic(err)
	}

	return doc
}

func (s *SpecgenPublicTestSuite) TestFull() {
	tests := []struct {
		name         string
		docs         []*openapi3.T
		validateFunc func([]byte, error)
	}{
		{
			name: "when groups are joined, each keeps its tag under its title, in order",
			docs: []*openapi3.T{
				spec[node]("zeta", "Zeta API"),
				spec[node]("alpha", "Alpha Management API"),
			},
			validateFunc: func(got []byte, err error) {
				s.Require().NoError(err)
				s.Contains(string(got), "# header\nopenapi: 3.0.0\ninfo:\n  title: osapi\n")
				var doc struct {
					Tags []struct {
						Name string `yaml:"name"`
					} `yaml:"tags"`
					Paths      yaml.Node `yaml:"paths"`
					Components struct {
						Schemas yaml.Node `yaml:"schemas"`
					} `yaml:"components"`
					Groups []struct {
						Name string   `yaml:"name"`
						Tags []string `yaml:"tags"`
					} `yaml:"x-tagGroups"`
					Areas []area.Area `yaml:"x-osapi-areas"`
				}
				s.Require().NoError(yaml.Unmarshal(got, &doc))
				s.Equal(area.All, doc.Areas, "the areas, with their labels, in order")
				s.Equal("Zeta_API_zeta_operations", doc.Tags[0].Name)
				s.Equal("Alpha_Management_API_alpha_operations", doc.Tags[1].Name)
				s.Equal("/api/zeta", doc.Paths.Content[0].Value, "in the order of the groups")
				s.Equal("Alpha Management API", doc.Groups[1].Name)
				s.Equal([]string{"Alpha_Management_API_alpha_operations"}, doc.Groups[1].Tags)
				s.Equal("ErrorResponse", doc.Components.Schemas.Content[0].Value, "first")
				s.Equal([]string{"type", "properties"}, []string{
					doc.Components.Schemas.Content[1].Content[0].Value,
					doc.Components.Schemas.Content[1].Content[2].Value,
				})
				s.Len(doc.Components.Schemas.Content, 4, "a shared schema is there once")
				s.Contains(string(got), "example: 134217728\n", "a whole number stays whole")
				s.Contains(string(got), "example: 2.5\n")
				s.Contains(
					string(got),
					"- Zeta_API_zeta_operations\n",
					"an operation has the new tag",
				)
			},
		},
		{
			name: "when two groups serve one path",
			docs: []*openapi3.T{spec[node]("zeta", "Zeta API"), spec[node]("zeta", "Again API")},
			validateFunc: func(_ []byte, err error) {
				s.EqualError(err, "path /api/zeta: Zeta API and Again API both define it")
			},
		},
		{
			name: "when two groups give one schema two shapes",
			docs: []*openapi3.T{spec[node]("zeta", "Zeta API"), func() *openapi3.T {
				doc := spec[node]("alpha", "Alpha API")
				doc.Components.Schemas["node"] = spec[other]("x", "X API").Components.Schemas["other"]

				return doc
			}()},
			validateFunc: func(_ []byte, err error) {
				s.EqualError(err, "schema node: Zeta API and Alpha API both define it")
			},
		},
		{
			name: "when a spec has no JSON",
			docs: []*openapi3.T{unencodable()},
			validateFunc: func(_ []byte, err error) {
				s.ErrorContains(err, "unsupported type")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(specgen.Full(tc.docs, "header"))
		})
	}
}

func (s *SpecgenPublicTestSuite) TestEncode() {
	got, err := specgen.Encode(spec[node]("zeta", "Zeta API"), "header")
	s.Require().NoError(err)
	s.Contains(string(got), "# header\ncomponents:\n")
	s.Contains(string(got), "example: 134217728\n")

	_, err = specgen.Encode(unencodable(), "header")
	s.ErrorContains(err, "unsupported type")
}

// unencodable is a spec whose extension JSON cannot encode.
func unencodable() *openapi3.T {
	doc := spec[node]("zeta", "Zeta API")
	doc.Extensions = map[string]any{"x-bad": make(chan int)}

	return doc
}

func TestSpecgenPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpecgenPublicTestSuite))
}
