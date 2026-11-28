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
	"context"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/apicontract"
)

// update rewrites the golden contract of the full spec:
//
//	go test ./internal/apicontract -run TestContractPublicTestSuite/TestJoinedSpec -update
var update = flag.Bool("update", false, "rewrite the golden contract of the full spec")

type ContractPublicTestSuite struct {
	suite.Suite

	doc map[string]any
}

func (s *ContractPublicTestSuite) load(
	text string,
) map[string]any {
	var doc map[string]any
	s.Require().NoError(yaml.Unmarshal([]byte(text), &doc))

	return doc
}

// TestJoinedSpec holds the full spec, which the controller serves and the
// clients come from, to its golden contract. A change to the API that is
// wanted updates the golden file in the same pull request, so the diff
// shows it; a change that is not wanted, such as a schema that a move to a
// generated spec changed, fails here.
func (s *ContractPublicTestSuite) TestJoinedSpec() {
	data, err := os.ReadFile(filepath.Join("..", "controller", "api", "gen", "api.yaml"))
	s.Require().NoError(err)

	got := strings.Join(apicontract.Lines(s.load(string(data))), "\n") + "\n"
	golden := filepath.Join("testdata", "joined.golden")
	if *update {
		s.Require().NoError(os.MkdirAll("testdata", 0o755))
		s.Require().NoError(os.WriteFile(golden, []byte(got), 0o644))
	}
	want, err := os.ReadFile(golden)
	s.Require().NoError(err)
	s.Equal(string(want), got)
}

// TestJoinedSpecIsValid holds the full spec to OpenAPI itself: a tool that
// validates a spec, as kin-openapi's router does, rejects one that breaks
// it, such as an object field whose example is a string.
func (s *ContractPublicTestSuite) TestJoinedSpecIsValid() {
	doc, err := openapi3.NewLoader().
		LoadFromFile(filepath.Join("..", "controller", "api", "gen", "api.yaml"))
	s.Require().NoError(err)
	s.NoError(doc.Validate(context.Background()))
}

func (s *ContractPublicTestSuite) TestLines() {
	tests := []struct {
		name string
		doc  string
		want []string
	}{
		{
			name: "when an operation has auth, parameters, a body and responses",
			doc: `
tags: [{name: agent}]
paths:
  /api/agent/{hostname}:
    parameters: [{name: ignored}]
    post:
      operationId: drainAgent
      tags: [agent]
      security: [{BearerAuth: ["agent:write"]}]
      parameters:
        - {$ref: '#/components/parameters/Hostname'}
        - {name: force, in: query, schema: {type: boolean}}
        - {$ref: '#/components/parameters/Missing'}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [reason]
              properties:
                reason: {type: string, enum: [a, b]}
                at: {type: string, format: date-time}
      responses:
        "200":
          content:
            application/json: {schema: {$ref: '#/components/schemas/Agent'}}
        "204": {description: none}
components:
  parameters:
    Hostname: {name: hostname, in: path, required: true, schema: {type: string}}
`,
			want: []string{
				"op POST /api/agent/{hostname} id=drainAgent tags=[agent] " +
					"auth=BearerAuth:agent:write " +
					"params=[path:hostname!:string query:force:boolean ref:Missing] " +
					"body=application/json:{at:string/date-time,reason!:string(a|b)}! " +
					"responses=[200=application/json:Agent 204=-]",
				"tag agent",
			},
		},
		{
			name: "when an operation has no security, it is public",
			doc: `
paths:
  /api/health:
    get:
      responses:
        "200":
          content:
            multipart/form-data: {schema: {type: array, items: {type: object, additionalProperties: {type: integer}}}}
`,
			want: []string{
				"op GET /api/health id= tags=[] auth=public params=[] body= " +
					"responses=[200=multipart/form-data:[map[integer]]]",
			},
		},
		{
			name: "when components have objects, refs and arrays",
			doc: `
components:
  schemas:
    Agent:
      type: object
      required: [name]
      properties:
        name: {type: string}
        labels: {type: object, additionalProperties: {type: string}}
        facts: {type: object, additionalProperties: true}
    Error: {$ref: '../components/gen/api.yaml#/components/schemas/ErrorResponse'}
    ErrorResponse: {$ref: '../components/gen/api.yaml#/components/schemas/ErrorResponse'}
    Names: {type: array, items: {type: string}}
`,
			want: []string{
				"schema Agent.facts = map[any]",
				"schema Agent.labels = map[string]",
				"schema Agent.name! = string",
				"schema Error = ErrorResponse",
				"schema Names = [string]",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, apicontract.Lines(s.load(tc.doc)))
		})
	}
}

func TestContractPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ContractPublicTestSuite))
}

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

func (s *ContractPublicTestSuite) SetupSuite() {
	data, err := os.ReadFile(filepath.Join("..", "controller", "api", "gen", "api.yaml"))
	s.Require().NoError(err)
	s.Require().NoError(yaml.Unmarshal(data, &s.doc))
}

// op is one operation of the full spec.
type op struct {
	verb, path string
	body       map[string]any
}

func (s *ContractPublicTestSuite) ops() []op {
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
func (s *ContractPublicTestSuite) TestConventions() {
	// Each operationId is Pascal case. A 202 is for work that is not done when
	// the answer goes, and each operation answers when it is done. A 201 has a
	// Location header, a 4xx or 5xx answers {"error": "..."}, a delete of the
	// controller answers 204, and a list answers items and total_items.
	schemas, _ := s.doc["components"].(map[string]any)["schemas"].(map[string]any)
	var bad []string
	for _, o := range s.ops() {
		where := o.verb + " " + o.path
		id, _ := o.body["operationId"].(string)
		if unicode.IsLower([]rune(id)[0]) {
			bad = append(bad, where+": operationId "+id+" is not Pascal case")
		}

		responses, _ := o.body["responses"].(map[string]any)
		node := strings.HasPrefix(o.path, "/api/node/")
		for code, r := range responses {
			switch {
			case code == "202":
				bad = append(bad, where+": answers 202")
			case code == "201":
				headers, _ := r.(map[string]any)["headers"].(map[string]any)
				if _, ok := headers["Location"]; !ok {
					bad = append(bad, where+": a 201 with no Location header")
				}
			case code[0] == '4' || code[0] == '5':
				if !probes[where+" "+code] && schema(r) != "ErrorResponse" {
					bad = append(bad, where+": a "+code+" that is not ErrorResponse")
				}
			}
		}
		if o.verb == "DELETE" && !node {
			if _, ok := responses["204"]; !ok {
				bad = append(bad, where+": a delete with no 204")
			}
		}
		if o.verb == "GET" && !node && !strings.Contains(o.path, "{") && !notCollections[o.path] {
			name := schema(responses["200"])
			props, _ := schemas[name].(map[string]any)["properties"].(map[string]any)
			for _, p := range []string{"items", "total_items"} {
				if _, ok := props[p]; !ok {
					bad = append(bad, where+": a list with no "+p)
				}
			}
		}
	}
	s.Empty(bad)
}
