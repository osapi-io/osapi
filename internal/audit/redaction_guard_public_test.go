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

package audit_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/audit"
)

// sensitiveLooking matches a property name that a reader would expect to carry
// a credential. It is deliberately broader than the deny list: the point is to
// stop a new field slipping past unreviewed, so a false positive costs one line
// in reviewedNonSensitive and a false negative costs a secret in the audit
// stream.
var sensitiveLooking = regexp.MustCompile(
	`(?i)pass|secret|token|cred|auth|key|bearer|signature|salt`,
)

// reviewedNonSensitive are property names that match sensitiveLooking and have
// been read and found to carry no secret. Each needs its reason, because the
// next person has to be able to tell a reviewed exception from an oversight.
var reviewedNonSensitive = map[string]string{
	"key": "node/user's is an SSH public key, which is not a secret; " +
		"node/sysctl's is a parameter name",
}

// RedactionGuardPublicTestSuite checks the audit deny list against what the
// OpenAPI specifications declare, rather than against what somebody remembered.
//
// internal/audit redacts by field name from a fixed list. Nothing otherwise
// fails when a domain is added with a credential-bearing field whose name is
// not on it, and the value would be written verbatim into the audit stream an
// admin reads, for that stream's whole retention. This turns that silent
// omission into a failing test at the moment the field is declared.
type RedactionGuardPublicTestSuite struct {
	suite.Suite
}

func (s *RedactionGuardPublicTestSuite) TestEverySensitiveLookingRequestFieldIsHandled() {
	specs := s.specFiles()
	s.NotEmpty(specs, "found no OpenAPI specifications to check")

	denied := audit.SensitiveFields()

	type finding struct{ spec, field string }

	var unhandled []finding

	for _, spec := range specs {
		for _, field := range s.requestBodyProperties(spec) {
			if !sensitiveLooking.MatchString(field) {
				continue
			}

			if denied[strings.ToLower(field)] {
				continue
			}

			if _, ok := reviewedNonSensitive[strings.ToLower(field)]; ok {
				continue
			}

			unhandled = append(unhandled, finding{spec: spec, field: field})
		}
	}

	for _, f := range unhandled {
		s.Failf("unhandled sensitive-looking request field",
			"%s declares a request body field %q, which looks like it carries a "+
				"credential and is neither redacted nor recorded as reviewed.\n\n"+
				"If it does carry a secret, add it to sensitiveFields in "+
				"internal/audit/summary.go.\n"+
				"If it does not, add it to reviewedNonSensitive in this file, with "+
				"the reason.",
			f.spec, f.field)
	}
}

// TestTheGuardWouldCatchANewField proves the check fails on the case it exists
// for. Without this, a guard that matched nothing would pass exactly as a
// working one does.
func (s *RedactionGuardPublicTestSuite) TestTheGuardWouldCatchANewField() {
	denied := audit.SensitiveFields()

	for _, field := range []string{
		"api_key", "passphrase", "credentials", "bearer", "auth_token",
		"client_secret", "refresh_token",
	} {
		s.True(sensitiveLooking.MatchString(field),
			"%q should read as sensitive to the guard", field)
		s.False(denied[field], "%q is not redacted today", field)
		s.NotContains(reviewedNonSensitive, field,
			"%q is not a reviewed exception, so the guard would fail on it", field)
	}
}

// TestEveryDeniedFieldIsStillLowercase guards the lookup itself: redactValue
// lowercases the incoming key before the map read, so an entry with any
// uppercase in it can never match.
func (s *RedactionGuardPublicTestSuite) TestEveryDeniedFieldIsStillLowercase() {
	for field := range audit.SensitiveFields() {
		s.Equal(strings.ToLower(field), field,
			"sensitiveFields key %q can never match, the lookup lowercases first",
			field)
	}
}

// specFiles returns each domain's own specification, skipping the combined one
// that redocly joins from them, so a field is reported against the file it is
// declared in.
func (s *RedactionGuardPublicTestSuite) specFiles() []string {
	var out []string

	root := filepath.Join("..", "controller", "api")
	combined := filepath.Join(root, "gen", "api.yaml")

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() != "api.yaml" || path == combined {
			return nil
		}

		out = append(out, path)

		return nil
	})
	s.Require().NoError(err)

	return out
}

// requestBodyProperties returns every property name reachable from any request
// body in one specification, at every depth, following local schema references.
func (s *RedactionGuardPublicTestSuite) requestBodyProperties(
	path string,
) []string {
	data, err := os.ReadFile(path) //nolint:gosec // a repository file, named by the walk above
	s.Require().NoError(err)

	var doc map[string]any
	s.Require().NoError(yaml.Unmarshal(data, &doc))

	schemas, _ := dig(doc, "components", "schemas").(map[string]any)

	seen := map[string]bool{}
	names := map[string]bool{}

	paths, _ := doc["paths"].(map[string]any)
	for _, item := range paths {
		operations, _ := item.(map[string]any)
		for _, op := range operations {
			operation, ok := op.(map[string]any)
			if !ok {
				continue
			}

			content, _ := dig(operation, "requestBody", "content").(map[string]any)
			for _, media := range content {
				m, ok := media.(map[string]any)
				if !ok {
					continue
				}

				collectProperties(m["schema"], schemas, seen, names)
			}
		}
	}

	out := make([]string, 0, len(names))
	for n := range names {
		out = append(out, n)
	}

	return out
}

// collectProperties walks a schema and records every property name it can
// reach, resolving a local $ref through the schemas map. seen stops a
// self-referential schema from recursing forever.
func collectProperties(
	schema any,
	schemas map[string]any,
	seen, names map[string]bool,
) {
	node, ok := schema.(map[string]any)
	if !ok {
		return
	}

	if ref, ok := node["$ref"].(string); ok {
		name := ref[strings.LastIndex(ref, "/")+1:]
		if seen[name] {
			return
		}

		seen[name] = true

		collectProperties(schemas[name], schemas, seen, names)

		return
	}

	if props, ok := node["properties"].(map[string]any); ok {
		for name, child := range props {
			names[name] = true

			collectProperties(child, schemas, seen, names)
		}
	}

	collectProperties(node["items"], schemas, seen, names)
	collectProperties(node["additionalProperties"], schemas, seen, names)

	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		list, _ := node[key].([]any)
		for _, child := range list {
			collectProperties(child, schemas, seen, names)
		}
	}
}

// dig walks nested maps by key, returning nil at the first key that is missing
// or is not a map.
func dig(
	m map[string]any,
	keys ...string,
) any {
	var current any = m

	for _, key := range keys {
		node, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		current = node[key]
	}

	return current
}

func TestRedactionGuardPublicTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(RedactionGuardPublicTestSuite))
}
