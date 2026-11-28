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
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// vaultIn is a behavior's input with a secret.Value field, under a name the
// fixed list does not have, and a map of them.
type vaultIn struct {
	Name       string                  `json:"name"`
	Passphrase secret.Value            `json:"passphrase"`
	Env        map[string]secret.Value `json:"env"`
}

type vaultOut struct{}

type vaultProvider struct{}

func unlock(
	_ context.Context,
	_ vaultProvider,
	_ vaultIn,
) (vaultOut, error) {
	return vaultOut{}, nil
}

var _ = behavior.Register(behavior.Define("vault", "Unlocks a vault", behavior.Node,
	behavior.Read("get", unlock),
))

// SummaryPublicTestSuite covers what an audit entry keeps of a request: enough to
// answer what ran, never a secret, never without bound.
type SummaryPublicTestSuite struct {
	suite.Suite
}

func (s *SummaryPublicTestSuite) TestSummarize() {
	tests := []struct {
		name         string
		body         string
		contentType  string
		validateFunc func(string)
	}{
		{
			name:        "a command and its arguments, which is the whole point",
			body:        `{"command":"systemctl","args":["restart","nginx"]}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Contains(got, "systemctl")
				s.Contains(got, "restart")
			},
		},
		{
			name:        "a password is not recorded",
			body:        `{"name":"deploy","password":"hunter2"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.NotContains(got, "hunter2")
				s.Contains(got, "[redacted]")
				s.Contains(got, "deploy")
			},
		},
		{
			name:        "a secret.Value field of a behavior is not recorded",
			body:        `{"name":"main","passphrase":"correct horse"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.NotContains(got, "correct horse")
				s.Contains(got, "[redacted]")
				s.Contains(got, "main")
			},
		},
		{
			name:        "a reference is recorded, so the entry shows which secret was used",
			body:        `{"passphrase":"secret://default/vault","password":"secret://default/pw"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Contains(got, `"passphrase":"secret://default/vault"`)
				s.Contains(got, `"password":"secret://default/pw"`)
			},
		},
		{
			name:        "a map of secrets keeps its names and references, and redacts each value",
			body:        `{"env":{"A":"hunter2","B":"secret://default/b","C":1}}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.NotContains(got, "hunter2")
				s.Contains(got, `"A":"[redacted]"`)
				s.Contains(got, `"B":"secret://default/b"`)
				s.Contains(got, `"C":"[redacted]"`)
			},
		},
		{
			name:        "a secret nested inside a list is not recorded either",
			body:        `{"users":[{"name":"a","password_hash":"$6$abc"}]}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.NotContains(got, "$6$abc")
				s.Contains(got, "[redacted]")
			},
		},
		{
			name:        "field names match whatever their case",
			body:        `{"Token":"abc123","Secret":"xyz"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.NotContains(got, "abc123")
				s.NotContains(got, "xyz")
			},
		},
		{
			name:        "an empty body summarizes to nothing",
			body:        "",
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Empty(got)
			},
		},
		{
			name:        "a body that is not json is described, not stored",
			body:        "--boundary\r\nContent-Disposition: form-data\r\n\r\nbinary",
			contentType: "multipart/form-data; boundary=boundary",
			validateFunc: func(got string) {
				s.Contains(got, "multipart/form-data")
				s.NotContains(got, "binary")
			},
		},
		{
			name:        "a body with no content type at all",
			body:        "something",
			contentType: "",
			validateFunc: func(got string) {
				s.Contains(got, "no content type")
			},
		},
		{
			name:        "json that does not parse",
			body:        `{"command":`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Contains(got, "unparsable json")
			},
		},
		{
			name:        "a long value is cut and says so",
			body:        `{"content_note":"` + strings.Repeat("x", 4000) + `"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Less(len(got), 2200)
				s.Contains(got, "[truncated]")
			},
		},
		{
			name:        "a cut that would land inside a character moves back",
			body:        `{"note":"` + strings.Repeat("é", 2000) + `"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Contains(got, "[truncated]")
				s.True(utf8.ValidString(got))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(audit.Summarize([]byte(tt.body), tt.contentType))
		})
	}
}

// TestSummarizeUnencodable covers the encoder failing, which a body that parsed
// cannot cause on its own.
func (s *SummaryPublicTestSuite) TestSummarizeUnencodable() {
	defer audit.ResetMarshalJSON()

	audit.SetMarshalJSON(func(any) ([]byte, error) {
		return nil, errors.New("encoder failed")
	})

	got := audit.Summarize([]byte(`{"command":"ls"}`), "application/json")
	s.Contains(got, "unsummarizable json")
	s.NotContains(got, "ls")
}

// TestJobIDSink covers the channel by which an entry learns its job ID without
// every handler reporting one.
func (s *SummaryPublicTestSuite) TestJobIDSink() {
	s.Run("a context with no sink records nothing and does not panic", func() {
		ctx := context.Background()
		audit.RecordJobID(ctx, "job-1")
		s.Empty(audit.JobIDFrom(ctx))
	})

	s.Run("the job a request created", func() {
		ctx := audit.WithJobIDSink(context.Background())
		audit.RecordJobID(ctx, "job-1")
		s.Equal("job-1", audit.JobIDFrom(ctx))
	})

	s.Run("the first job wins", func() {
		ctx := audit.WithJobIDSink(context.Background())
		audit.RecordJobID(ctx, "job-1")
		audit.RecordJobID(ctx, "job-2")
		s.Equal("job-1", audit.JobIDFrom(ctx))
	})

	s.Run("an empty job id is not a job", func() {
		ctx := audit.WithJobIDSink(context.Background())
		audit.RecordJobID(ctx, "")
		s.Empty(audit.JobIDFrom(ctx))
	})

	s.Run("concurrent dispatches do not race", func() {
		ctx := audit.WithJobIDSink(context.Background())

		var wg sync.WaitGroup

		wg.Add(4)

		for i := range 4 {
			go func(n int) {
				defer wg.Done()

				audit.RecordJobID(ctx, string(rune('a'+n)))
				_ = audit.JobIDFrom(ctx)
			}(i)
		}

		wg.Wait()

		s.NotEmpty(audit.JobIDFrom(ctx))
	})
}

// TestStepsSink covers the channel by which an entry learns the chain of the
// job of its request.
func (s *SummaryPublicTestSuite) TestStepsSink() {
	update := []types.Step{{Behavior: "managedfile", Method: "update"}}
	failed := []types.Step{{Behavior: "managedfile", Method: "update", Error: "denied"}}

	tests := []struct {
		name   string
		ctx    func() context.Context
		record func(context.Context)
		want   []audit.HostSteps
	}{
		{
			name:   "a context with no sink records nothing and does not panic",
			ctx:    context.Background,
			record: func(ctx context.Context) { audit.RecordSteps(ctx, "job-1", "web-01", update) },
		},
		{
			name: "the steps of each host, sorted by hostname, also when one failed",
			ctx:  func() context.Context { return audit.WithJobIDSink(context.Background()) },
			record: func(ctx context.Context) {
				audit.RecordJobID(ctx, "job-1")
				audit.RecordSteps(ctx, "job-1", "web-02", failed)
				audit.RecordSteps(ctx, "job-1", "web-01", update)
			},
			want: []audit.HostSteps{
				{Hostname: "web-01", Steps: update},
				{Hostname: "web-02", Steps: failed},
			},
		},
		{
			name: "a host that made no call is left out",
			ctx:  func() context.Context { return audit.WithJobIDSink(context.Background()) },
			record: func(ctx context.Context) {
				audit.RecordJobID(ctx, "job-1")
				audit.RecordSteps(ctx, "job-1", "web-01", nil)
			},
		},
		{
			name: "the steps of a job that is not the job of the request",
			ctx:  func() context.Context { return audit.WithJobIDSink(context.Background()) },
			record: func(ctx context.Context) {
				audit.RecordJobID(ctx, "job-1")
				audit.RecordSteps(ctx, "job-2", "web-01", update)
			},
		},
		{
			name: "the steps before a job is recorded",
			ctx:  func() context.Context { return audit.WithJobIDSink(context.Background()) },
			record: func(ctx context.Context) {
				audit.RecordSteps(ctx, "job-1", "web-01", update)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := tt.ctx()
			tt.record(ctx)
			s.Equal(tt.want, audit.StepsFrom(ctx))
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestSummaryPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SummaryPublicTestSuite))
}

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
	"public_key": "agent/join's is the agent's Ed25519 public key, which the " +
		"admission record keeps and shows; not a secret",
	"encryption_key": "agent/join's is the agent's P-256 public key; not a secret",
	"signature": "agent/join's is the agent's signature over its own request, " +
		"which proves it holds its key; it is no credential and cannot be replayed " +
		"as one",
	"key": "node/user's is an SSH public key, which is not a secret; " +
		"node/sysctl's is a parameter name",
}

func (s *SummaryPublicTestSuite) TestEverySensitiveLookingRequestFieldIsHandled() {
	specs := s.specFiles()
	s.NotEmpty(specs)

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

	// A request body field that looks like it carries a credential is
	// redacted or reviewed. If it carries a secret, add it to sensitiveFields
	// in internal/audit/summary.go. If not, add it to reviewedNonSensitive in
	// this file, with the reason.
	s.Empty(unhandled)
}

// TestTheGuardWouldCatchANewField proves the check fails on the case it exists
// for. Without this, a guard that matched nothing would pass exactly as a
// working one does.
func (s *SummaryPublicTestSuite) TestTheGuardWouldCatchANewField() {
	denied := audit.SensitiveFields()

	for _, field := range []string{
		"api_key", "passphrase", "credentials", "bearer", "auth_token",
		"client_secret", "refresh_token",
	} {
		s.True(sensitiveLooking.MatchString(field))
		s.False(denied[field])
		s.NotContains(reviewedNonSensitive, field)
	}
}

// TestEveryDeniedFieldIsStillLowercase guards the lookup itself: redactValue
// lowercases the incoming key before the map read, so an entry with any
// uppercase in it can never match.
func (s *SummaryPublicTestSuite) TestEveryDeniedFieldIsStillLowercase() {
	for field := range audit.SensitiveFields() {
		s.Equal(strings.ToLower(field), field)
	}
}

// specFiles returns the full specification of the API, which has the
// request body of each endpoint group and behavior.
func (s *SummaryPublicTestSuite) specFiles() []string {
	return []string{filepath.Join("..", "controller", "api", "gen", "api.yaml")}
}

// requestBodyProperties returns every property name reachable from any request
// body in one specification, at every depth, following local schema references.
func (s *SummaryPublicTestSuite) requestBodyProperties(
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

func TestRedactionGuardPublicTestSuite(
	t *testing.T,
) {
	t.Parallel()
	suite.Run(t, new(SummaryPublicTestSuite))
}
