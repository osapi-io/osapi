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
package apicontract_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

// probes are the operations whose failure has a body of its own, not an
// ErrorResponse: a readiness probe answers 503 and says what is not ready.
var probes = map[string]bool{"GET /api/health/ready 503": true}

// notCollections are the controller's GETs without a key that are not a
// list of things: the health of the controller, the catalog, one document
// that the UI reads, and the controller's key and the NATS server's certificate, with their rotations.
var notCollections = map[string]bool{
	"/api/health": true, "/api/health/ready": true, "/api/health/status": true,
	"/api/catalog": true, "/api/pki/key": true, "/api/pki/nats": true,
}

type ConventionsPublicTestSuite struct {
	suite.Suite

	doc map[string]any
}

func (s *ConventionsPublicTestSuite) SetupSuite() {
	data, err := os.ReadFile(filepath.Join("..", "controller", "api", "gen", "api.yaml"))
	s.Require().NoError(err)
	s.Require().NoError(yaml.Unmarshal(data, &s.doc))
}

// op is one operation of the full spec.
type op struct {
	verb, path string
	body       map[string]any
}

func (s *ConventionsPublicTestSuite) ops() []op {
	var out []op
	paths, _ := s.doc["paths"].(map[string]any)
	for path, item := range paths {
		verbs, _ := item.(map[string]any)
		for verb, body := range verbs {
			b, ok := body.(map[string]any)
			if !ok || b["operationId"] == nil {
				continue
			}
			out = append(out, op{verb: strings.ToUpper(verb), path: path, body: b})
		}
	}
	sort.Slice(
		out,
		func(i, j int) bool { return out[i].path+out[i].verb < out[j].path+out[j].verb },
	)

	return out
}

// schema is the name of the JSON schema a response refers to, or "".
func schema(
	response any,
) string {
	r, _ := response.(map[string]any)
	content, _ := r["content"].(map[string]any)
	js, _ := content["application/json"].(map[string]any)
	sch, _ := js["schema"].(map[string]any)
	ref, _ := sch["$ref"].(string)

	return ref[strings.LastIndex(ref, "/")+1:]
}

// TestConventions holds the API to the conventions of
// components/osapi/api.md in osapi-io/specs. A failure names the operation
// and the rule; change the operation, not the test.
func (s *ConventionsPublicTestSuite) TestConventions() {
	schemas, _ := s.doc["components"].(map[string]any)["schemas"].(map[string]any)
	for _, o := range s.ops() {
		where := o.verb + " " + o.path
		id, _ := o.body["operationId"].(string)
		s.Falsef(unicode.IsLower([]rune(id)[0]),
			"%s: operationId %s starts with a small letter; each is Pascal case", where, id)

		responses, _ := o.body["responses"].(map[string]any)
		node := strings.HasPrefix(o.path, "/api/node/")
		for code, r := range responses {
			switch {
			case code == "202":
				s.Failf(
					"async",
					"%s: 202 is for work that is not done when the answer goes; each operation answers when it is done",
					where,
				)
			case code == "201":
				headers, _ := r.(map[string]any)["headers"].(map[string]any)
				s.Containsf(
					headers,
					"Location",
					"%s: a 201 has a Location header that names what it created",
					where,
				)
			case code[0] == '4' || code[0] == '5':
				if !probes[where+" "+code] {
					s.Equalf(
						"ErrorResponse",
						schema(r),
						"%s: a %s answers {\"error\": \"...\"}",
						where,
						code,
					)
				}
			}
		}
		if o.verb == "DELETE" && !node {
			s.Containsf(responses, "204", "%s: a delete of the controller answers 204", where)
		}
		if o.verb == "GET" && !node && !strings.Contains(o.path, "{") && !notCollections[o.path] {
			name := schema(responses["200"])
			props, _ := schemas[name].(map[string]any)["properties"].(map[string]any)
			s.Containsf(props, "items", "%s: a list answers items and total_items", where)
			s.Containsf(props, "total_items", "%s: a list answers items and total_items", where)
		}
	}
}

func TestConventionsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ConventionsPublicTestSuite))
}
