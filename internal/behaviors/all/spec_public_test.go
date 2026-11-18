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
package all_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

// repoRoot is the repository root, from this package's directory.
const repoRoot = "../../.."

// migrating pairs each behavior that replaces a hand-written domain with that
// domain's spec, from the repository root: the behavior's generated spec, and
// the hand-written one. Until a domain's cutover, the behavior's spec must
// describe the same API apart from the /next prefix, its prose and the
// constraints the old spec omitted. Add a row when a behavior is built to
// replace a domain; delete it with the hand-written spec at the cutover. When
// the list is empty the migration is done, and this test goes too.
//
// sysctl has no row: its reads were changed on purpose before its cutover, to
// report every parameter rather than only the ones osapi manages, so its API
// no longer matches the hand-written one.
var migrating = []struct {
	generated string
	spec      string
}{}

type SpecPublicTestSuite struct {
	suite.Suite
}

// TestSpec loads every spec the behaviors' endpoint groups produced, each
// behavior's gen/api.yaml or gen/api.next.yaml, with kin-openapi, the library
// oapi-codegen reads specs with, and validates it, references to the shared
// spec included.
func (s *SpecPublicTestSuite) TestSpec() {
	var specs []string
	for _, pattern := range []string{
		"internal/behaviors/*/gen/api*.yaml",
		"internal/behaviors/*/*/gen/api*.yaml",
		"internal/controller/api/next/gen/api.yaml",
	} {
		found, err := filepath.Glob(filepath.Join(repoRoot, pattern))
		s.Require().NoError(err)
		specs = append(specs, found...)
	}
	s.Require().NotEmpty(specs)

	for _, spec := range specs {
		s.Run(spec, func() {
			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = true
			doc, err := loader.LoadFromFile(spec)
			s.Require().NoError(err)
			s.NoError(doc.Validate(context.Background()))
		})
	}
}

// TestParity holds each behavior that replaces a hand-written domain to the
// same API as that domain's spec, apart from the /next prefix, its prose and
// the constraints the old spec omitted.
func (s *SpecPublicTestSuite) TestParity() {
	parts := []struct {
		name string
		part func(doc map[string]any) []string
	}{
		{name: "operations", part: operations},
		{name: "schemas", part: schemas},
		{name: "parameters", part: parameters},
	}

	for _, m := range migrating {
		s.Run("when "+m.generated+" replaces "+m.spec, func() {
			var want, got map[string]any
			for path, doc := range map[string]*map[string]any{m.spec: &want, m.generated: &got} {
				data, err := os.ReadFile(filepath.Join(repoRoot, path))
				s.Require().NoError(err)
				s.Require().NoError(yaml.Unmarshal(data, doc))
			}
			for _, p := range parts {
				s.Equal(p.part(want), p.part(got), "%s: %s", m.generated, p.name)
			}
		})
	}
}

// operations is one line per operation: path without /next, verb, id, scopes,
// parameters, request body and response codes.
func operations(
	doc map[string]any,
) []string {
	var out []string
	for path, item := range doc["paths"].(map[string]any) {
		path = strings.Replace(path, "/api/next/", "/api/", 1)
		for verb, raw := range item.(map[string]any) {
			op := raw.(map[string]any)
			var scopes []string
			for _, sec := range op["security"].([]any) {
				for _, list := range sec.(map[string]any) {
					for _, sc := range list.([]any) {
						scopes = append(scopes, sc.(string))
					}
				}
			}
			var params []string
			for _, p := range op["parameters"].([]any) {
				params = append(params, ref(p))
			}
			body := ""
			if rb, ok := op["requestBody"].(map[string]any); ok {
				body = ref(
					rb["content"].(map[string]any)["application/json"].(map[string]any)["schema"],
				)
			}
			var codes []string
			for code, r := range op["responses"].(map[string]any) {
				schema := r.(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"]
				codes = append(codes, code+"="+ref(schema))
			}
			sort.Strings(codes)
			out = append(out, fmt.Sprintf("%s %s %s scopes=%v params=%v body=%s responses=%v",
				verb, path, op["operationId"], scopes, params, body, codes))
		}
	}
	sort.Strings(out)

	return out
}

// schemas is one line per schema property: its type, enum and whether it is
// required.
func schemas(
	doc map[string]any,
) []string {
	var out []string
	for name, raw := range doc["components"].(map[string]any)["schemas"].(map[string]any) {
		schema := raw.(map[string]any)
		if _, ok := schema["$ref"]; ok {
			out = append(out, name+" = "+ref(schema))

			continue
		}
		required := map[string]bool{}
		for _, r := range asSlice(schema["required"]) {
			required[r.(string)] = true
		}
		for prop, p := range schema["properties"].(map[string]any) {
			pm := p.(map[string]any)
			out = append(out, fmt.Sprintf("%s.%s type=%v format=%v enum=%v items=%s required=%t",
				name, prop, pm["type"], pm["format"], pm["enum"], ref(pm["items"]), required[prop]))
		}
	}
	sort.Strings(out)

	return out
}

// parameters is one line per path parameter: its name, place and pattern.
func parameters(
	doc map[string]any,
) []string {
	params := doc["components"].(map[string]any)["parameters"].(map[string]any)
	out := make([]string, 0, len(params))
	for name, raw := range params {
		p := raw.(map[string]any)
		schema := p["schema"].(map[string]any)
		out = append(out, fmt.Sprintf("%s name=%v in=%v required=%v type=%v",
			name, p["name"], p["in"], p["required"], schema["type"]))
	}
	sort.Strings(out)

	return out
}

// ref is the schema or parameter name a $ref points at, whichever file it is
// in.
func ref(
	v any,
) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	r, _ := m["$ref"].(string)

	return r[strings.LastIndex(r, "/")+1:]
}

func asSlice(
	v any,
) []any {
	s, _ := v.([]any)

	return s
}

func TestSpecPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpecPublicTestSuite))
}
