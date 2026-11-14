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

package openapi_test

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

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/openapi"
	_ "github.com/osapi-io/osapi/internal/validation"
)

// repoRoot is the repository root, from this package's directory.
const repoRoot = "../../.."

// migrating pairs each behavior that replaces a hand-written domain with that
// domain's spec, from the repository root. Until a domain's cutover, the
// behavior's generated spec must describe the same API apart from the /next
// prefix, its prose and the constraints the old spec omitted. Add a row when a
// behavior is built to replace a domain; delete it with the hand-written spec
// at the cutover. When the list is empty the migration is done, and this test
// goes too.
//
// sysctl has no row: its reads were changed on purpose before its cutover, to
// report every parameter rather than only the ones osapi manages, so its API
// no longer matches the hand-written one.
var migrating = []struct {
	behavior *behavior.Behavior
	spec     string
}{}

type OpenAPIPublicTestSuite struct {
	suite.Suite
}

func (s *OpenAPIPublicTestSuite) SetupSuite() {
	behavior.ValidatorSchema("spec_key", behavior.Pattern("^[a-z]+$"))
}

type specProvider struct{}

type specListIn struct {
	Prefix string `json:"prefix" doc:"Only keys with this prefix"`
	Limit  int    `json:"limit"                                   validate:"min=1,max=100"`
}

type specRich struct {
	Key   string            `json:"key"   validate:"required,spec_key,len=3" doc:"The key"`
	Mode  string            `json:"mode"  validate:"oneof=a b"`
	Addr  string            `json:"addr"  validate:"ip"`
	Tags  []string          `json:"tags"  validate:"min=1,max=3"`
	Attrs map[string]string `json:"attrs"`
	Ratio float64           `json:"ratio" validate:"max=1"`
	Token string            `json:"token"                                                  secret:"true"`
}

type specKey struct {
	Key string `json:"key" validate:"required" doc:"The key"`
}

type specBadTagIn struct {
	X string `json:"x" validate:"not_registered"`
}

type specEntry struct {
	Key   string `json:"key"   doc:"The key"`
	Value string `json:"value" doc:"The value"`
}

type specResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
}

type specReservedOut struct {
	Hostname string `json:"hostname"`
	Value    string `json:"value"`
}

func specList(_ context.Context, _ specProvider, _ specListIn) ([]specEntry, error) {
	return nil, nil
}

func specCreate(_ context.Context, _ specProvider, _ specRich) (specResult, error) {
	return specResult{}, nil
}

func specDelete(_ context.Context, _ specProvider, _ specKey) (specResult, error) {
	return specResult{}, nil
}

func specGet(_ context.Context, _ specProvider, _ specKey) (specEntry, error) {
	return specEntry{}, nil
}

func specReserved(_ context.Context, _ specProvider, _ specKey) (specReservedOut, error) {
	return specReservedOut{}, nil
}

func specBadTag(_ context.Context, _ specProvider, _ specBadTagIn) (specEntry, error) {
	return specEntry{}, nil
}

func specBadOut(_ context.Context, _ specProvider, _ specKey) (chan int, error) {
	return nil, nil
}

// spec renders bs and parses the result back, so a test asserts on the
// document's structure rather than its text.
func (s *OpenAPIPublicTestSuite) spec(
	bs ...*behavior.Behavior,
) map[string]any {
	out, err := openapi.OpenAPI(bs)
	s.Require().NoError(err)

	var doc map[string]any
	s.Require().NoError(yaml.Unmarshal(out, &doc))

	return doc
}

func dig(
	v any,
	path ...string,
) any {
	for _, p := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[p]
	}

	return v
}

func (s *OpenAPIPublicTestSuite) TestOpenAPI() {
	full := behavior.Define("kv", "Key values", behavior.Node, behavior.API,
		behavior.Read("list", specList).Describe("List them"),
		behavior.Write("create", specCreate).Describe("Create one").
			Explain("Writes the key and reports whether it changed."),
		behavior.Write("update", specCreate).Key("key").Describe("Update one"),
		behavior.Write("delete", specDelete).Key("key").Describe("Delete one"),
		behavior.Read("status", specReserved).Key("key").Describe("Status"),
	)
	grouped := behavior.Define(
		"dns",
		"Resolvers",
		behavior.Node,
		behavior.API,
		behavior.Under("network"),
		behavior.Read("get", specGet).Key("key").Describe("Get one"),
	)
	hidden := behavior.Define("inner", "Not served", behavior.Node,
		behavior.Read("get", specGet).Describe("Get"),
	)

	tests := []struct {
		name         string
		behaviors    []*behavior.Behavior
		validateFunc func(doc map[string]any)
	}{
		{
			name:      "when a behavior is served, its paths, operations and schemas are generated",
			behaviors: []*behavior.Behavior{full},
			validateFunc: func(doc map[string]any) {
				list := dig(doc, "paths", "/api/next/node/{hostname}/kv", "get")
				s.Equal("GetNodeKv", dig(list, "operationId"))
				params := dig(list, "parameters").([]any)
				s.Len(params, 3, "hostname, then the two inputs as query parameters")
				s.Equal("prefix", dig(params[1], "name"))
				s.Equal("query", dig(params[1], "in"))
				s.Nil(dig(list, "responses", "404"), "a list has no key to miss")
				s.Nil(dig(list, "description"), "an operation without Explain has no description")
				s.Equal("Each target host's result.", dig(list, "responses", "200", "description"))

				create := dig(doc, "paths", "/api/next/node/{hostname}/kv", "post")
				s.Equal("PostNodeKv", dig(create, "operationId"))
				s.Equal("Create one", dig(create, "summary"))
				s.Equal(
					"Writes the key and reports whether it changed.",
					dig(create, "description"),
				)
				s.Equal("Input to kv create.", dig(create, "requestBody", "description"))
				s.Equal("Each target host's result, with whether it changed.",
					dig(create, "responses", "200", "description"))
				s.Equal("#/components/schemas/KvCreateRequest",
					dig(create, "requestBody", "content", "application/json", "schema", "$ref"))

				del := dig(doc, "paths", "/api/next/node/{hostname}/kv/{key}", "delete")
				s.Nil(dig(del, "requestBody"), "the key is in the path, so there is no body")
				s.Equal(
					"The host has no kv with that key.",
					dig(del, "responses", "404", "description"),
				)
				s.Equal("The job could not be run or its result could not be read.",
					dig(del, "responses", "500", "description"))

				req := dig(doc, "components", "schemas", "KvCreateRequest", "properties")
				s.Equal("^[a-z]+$", dig(req, "key", "pattern"))
				s.Equal(3, dig(req, "key", "minLength"))
				s.Equal([]any{"a", "b"}, dig(req, "mode", "enum"))
				s.Equal("ip", dig(req, "addr", "format"))
				s.Equal("password", dig(req, "token", "format"))
				s.Equal(true, dig(req, "token", "writeOnly"))
				s.Equal(1, dig(req, "tags", "minItems"))
				s.Equal(3, dig(req, "tags", "maxItems"))
				s.Equal("string", dig(req, "attrs", "additionalProperties", "type"))
				s.EqualValues(1, dig(req, "ratio", "maximum"))

				s.NotNil(dig(doc, "components", "schemas", "KvCollectionResponse"))
				// The row is named for the behavior and the output's Go type.
				row := dig(doc, "components", "schemas", "KvspecEntry", "properties")
				s.NotNil(dig(row, "hostname"))
				s.NotNil(dig(row, "error"))
				s.Equal("Key values", dig(doc, "tags").([]any)[0].(map[string]any)["description"])
			},
		},
		{
			name:      "when a behavior is grouped, its names and tag carry the group",
			behaviors: []*behavior.Behavior{grouped},
			validateFunc: func(doc map[string]any) {
				op := dig(doc, "paths", "/api/next/node/{hostname}/network/dns/{key}", "get")
				s.Equal("GetNodeNetworkDnsByKey", dig(op, "operationId"))
				s.Equal([]any{"network_dns_operations"}, dig(op, "tags"))
				s.NotNil(dig(doc, "components", "parameters", "NetworkDnsKey"))
			},
		},
		{
			name:      "when a behavior has no API, it is left out",
			behaviors: []*behavior.Behavior{hidden},
			validateFunc: func(doc map[string]any) {
				s.Empty(dig(doc, "paths"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(s.spec(tc.behaviors...))
		})
	}
}

func (s *OpenAPIPublicTestSuite) TestOpenAPIErrors() {
	tests := []struct {
		name string
		b    *behavior.Behavior
	}{
		{
			name: "when an input has no schema",
			b: behavior.Define("x", "x", behavior.Node, behavior.API,
				behavior.Read("get", specBadTag).Describe("Get")),
		},
		{
			name: "when the key is not an input field",
			b: behavior.Define("x", "x", behavior.Node, behavior.API,
				behavior.Read("get", specGet).Key("nope").Describe("Get")),
		},
		{
			name: "when an output has no schema",
			b: behavior.Define("x", "x", behavior.Node, behavior.API,
				behavior.Read("get", specBadOut).Describe("Get")),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, err := openapi.OpenAPI([]*behavior.Behavior{tc.b})
			s.Error(err)
		})
	}
}

// TestOpenAPISpec loads the spec the generator wrote with kin-openapi, the
// library oapi-codegen reads specs with, and validates it.
func (s *OpenAPIPublicTestSuite) TestOpenAPISpec() {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(filepath.Join(repoRoot, behavior.SpecPath))
	s.Require().NoError(err)
	s.NoError(doc.Validate(context.Background()))
}

// TestOpenAPIParity holds each behavior that replaces a hand-written domain
// to the same API as that domain's spec, apart from the /next prefix, its
// prose and the constraints the old spec omitted.
func (s *OpenAPIPublicTestSuite) TestOpenAPIParity() {
	parts := []struct {
		name string
		part func(doc map[string]any) []string
	}{
		{name: "operations", part: operations},
		{name: "schemas", part: schemas},
		{name: "parameters", part: parameters},
	}

	for _, m := range migrating {
		s.Run("when "+m.behavior.Name+" replaces "+m.spec, func() {
			hand, err := os.ReadFile(filepath.Join(repoRoot, m.spec))
			s.Require().NoError(err)

			gen := s.spec(m.behavior)
			var want map[string]any
			s.Require().NoError(yaml.Unmarshal(hand, &want))
			for _, p := range parts {
				s.Equal(p.part(want), p.part(gen), "%s: %s", m.behavior.Name, p.name)
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

func TestOpenAPIPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OpenAPIPublicTestSuite))
}
